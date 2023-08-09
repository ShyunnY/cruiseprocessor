package cruiseprocessor

import (
	"context"
	"github.com/ShyunnY/cruiseprocessor/internal/metadata"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
)

// NewFactory new processor factory
func NewFactory() processor.Factory {
	// set about processor metadata
	return processor.NewFactory(
		metadata.Type,
		createDefaultConfig,
		processor.WithTraces(createTracesProcessor, metadata.TracesStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createTracesProcessor(
	ctx context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Traces,
) (processor.Traces, error) {
	return newCruiseProcessor(
		ctx,
		set,
		cfg,
		nextConsumer,
	)
}
