package cruiseprocessor

import (
	"context"
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor/processortest"
	"math/rand"
	"testing"
	"time"
)

func generateTraceData(duration int64) ptrace.Traces {
	low := rand.Uint64()
	high := rand.Uint64()
	td := ptrace.NewTraces()
	rs := td.ResourceSpans().AppendEmpty()
	span := rs.ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	span.SetTraceID(UInt64ToTraceID(low, high))
	span.SetStartTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	span.SetEndTimestamp(pcommon.NewTimestampFromTime(time.Now().Add(time.Millisecond * time.Duration(duration))))
	return td
}
func UInt64ToTraceID(high, low uint64) pcommon.TraceID {
	traceID := [16]byte{}
	binary.BigEndian.PutUint64(traceID[:8], high)
	binary.BigEndian.PutUint64(traceID[8:], low)
	return traceID
}
func TestCruise_Duration(t *testing.T) {

	factory := NewFactory()

	cfg := factory.CreateDefaultConfig()
	oCfg := cfg.(*Config)
	oCfg.Duration = 200

	tp, err := factory.CreateTracesProcessor(context.TODO(), processortest.NewNopCreateSettings(), oCfg, consumertest.NewNop())
	assert.NoError(t, err)
	assert.NotNil(t, tp)

	data := generateTraceData(100)
	err = tp.ConsumeTraces(context.TODO(), data)
	assert.NoError(t, err)

}

func TestCruise_Sampling(t *testing.T) {
	factory := NewFactory()
	const num = 100
	count := 0

	cfg := factory.CreateDefaultConfig()
	oCfg := cfg.(*Config)
	oCfg.Duration = 200
	oCfg.Sample = Sampling{
		Percentage: 100,
		//Normal:     true,
		//Salt:       "123",
	}
	sink := new(consumertest.TracesSink)
	tp, err := factory.CreateTracesProcessor(
		context.TODO(),
		processortest.NewNopCreateSettings(),
		oCfg,
		sink,
	)

	assert.NoError(t, err)
	assert.NotNil(t, tp)
	time.Sleep(time.Second)

	for i := 0; i < num; i++ {
		err = tp.ConsumeTraces(context.TODO(), generateTraceData(10))
		assert.NoError(t, err)
	}
	time.Sleep(time.Second * 3)
	for _, traces := range sink.AllTraces() {
		rss := traces.ResourceSpans()
		for i := 0; i < rss.Len(); i++ {
			ssp := rss.At(i).ScopeSpans()

			for j := 0; j < ssp.Len(); j++ {
				sps := ssp.At(j)
				sp := sps.Spans()
				for k := 0; k < sp.Len(); k++ {
					_ = sp.At(k)
					count++
				}
			}
		}
	}

	assert.NotZero(t, count)
}
