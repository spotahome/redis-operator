package main

import (
	"context"
	"fmt"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

type fakeService struct {
	tracer opentracing.Tracer
}

func (f *fakeService) makeOperation(ctx context.Context, opName string, duration time.Duration, failRandomly bool) (context.Context, error) {

	// Create the span.
	pSpan := opentracing.SpanFromContext(ctx)
	span := f.tracer.StartSpan(opName, opentracing.ChildOf(pSpan.Context()))
	newCtx := opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	// Fail sometimes (crappy and fast version).
	var err error
	if time.Now().Nanosecond()%7 == 0 {
		ext.Error.Set(span, true)
		err = fmt.Errorf("randomly failed")
	}

	time.Sleep(duration)
	return newCtx, err
}
