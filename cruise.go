package cruiseprocessor

import (
	"context"
	"github.com/google/uuid"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
)

const (
	normal = iota
	exception
)

type cruiseProcessor struct {
	config Config

	// private sample rate for use compute sample
	sampleRate uint64
	salt       string

	pool         *WorkerPool
	logger       *zap.Logger
	nextConsumer consumer.Traces
	ctx          context.Context
}

func newCruiseProcessor(ctx context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Traces) (processor.Traces, error) {

	config := checkConfig(cfg.(*Config))

	cp := &cruiseProcessor{
		config: *config,

		// use internal compute sample rate
		sampleRate:   computeSampleRate(config.Sample.Percentage),
		logger:       set.Logger,
		pool:         NewWorkerPool(defaultWorkerNum),
		nextConsumer: nextConsumer,
		ctx:          ctx,
	}

	return cp, nil
}

func (cp *cruiseProcessor) Start(ctx context.Context, host component.Host) error {
	// start work pool
	// the passed context should not be used, as it will end quickly
	go func() {
		if err := cp.pool.Run(context.Background()); err != nil {
			// TODO: consider how shutdown cruise processor
		}
	}()
	return nil
}

func (cp *cruiseProcessor) Shutdown(ctx context.Context) error {
	// TODO: when shutdown is called, consider completing all worker pools
	return nil
}

func (cp *cruiseProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

func (cp *cruiseProcessor) ConsumeTraces(ctx context.Context, td ptrace.Traces) (err error) {
	// TODO : 这里需要补充编号进行轮询分配
	cp.pool.Execute(1, func() error {
		td.ResourceSpans().RemoveIf(func(rss ptrace.ResourceSpans) bool {
			rss.ScopeSpans().RemoveIf(func(sps ptrace.ScopeSpans) bool {
				sps.Spans().RemoveIf(func(sp ptrace.Span) bool {

					traceID := sp.TraceID()
					id := uuid.New().String()
					var errFlag int
					// head sample for all span
					if cp.config.Sample.All && !hasSampling([]byte(id), cp.salt, cp.sampleRate) {
						return true
					}

					// handle according to the processor list
					// todo: more processors will be added in the future
					errFlag = cp.processDuration(sp)

					// tail sample for normal span
					if cp.config.Sample.Normal && errFlag == 0 {
						if !hasSampling(traceID[:], cp.salt, cp.sampleRate) {
							return true
						}
					}

					// by default, the error trace will not enter the execute hasSampling()
					// we want to capture all error traces
					return false
				})

				return sps.Spans().Len() == 0
			})
			return rss.ScopeSpans().Len() == 0
		})
		return cp.nextConsumer.ConsumeTraces(ctx, td)
	})
	return err
}

func checkConfig(cfg *Config) *Config {
	// check sample
	if !cfg.Sample.All && !cfg.Sample.Normal {
		// default all sample
		cfg.Sample.All = true
	}

	if cfg.Sample.Salt == "" {
		// use default salt val of hostname
		cfg.Sample.Salt = defaultSaltVal()
	}

	if cfg.Sample.Percentage >= 100 {
		cfg.Sample.Percentage = 100
		cfg.Sample.All = true
	}
	return cfg
}

func (cp *cruiseProcessor) processDuration(span ptrace.Span) int {
	// we think that full sampling is required
	if cp.config.Duration == 0 {
		return normal
	}

	startTs := span.StartTimestamp()
	endTs := span.EndTimestamp()

	subTime := endTs.AsTime().Sub(startTs.AsTime())
	beyond := subTime.Milliseconds() > cp.config.Duration

	if beyond {
		// traces staining
		// TODO: consider adding traffic staining functionality to your attributed customize
		span.Attributes().PutStr("health.latency", "true")
		return exception
	}
	return normal
}
