# Tracing

Kooper has builting tracing support using [Opentracing][opentracing-url]. This will allow a controller/operator trace all the process from the moment kooper controller starts handling the event until it finishes handling it. The handlers of the objects will receive a `context` where the parent span (coming from Kooper) can be retrieved and continue the trace.

## Root span

Kooper will create a trace for each object received using the `Retriever` (using resync/list or watch events). This will create a root span that will set important data.

### Relevant tags

* `kubernetes.object.namespace`: The namespace of the object being processed.
* `kubernetes.object.name`: the name of the object being processed.
* `kubernetes.object.key`: Key of the object, with namespace/objectName style.
* `component`: Always kooper.
* `kooper.controller`: The name of the controller (got from the configuration).
* `controller.cfg.*`: Some tags with the configuration of the controller.
* `kubernetes.object.total_processed_times`: the number of times the same object has been processed (1 + retries).
* `kubernetes.object.processing_retry`: Marks if this processing object is a retry of a previous processed one (means that the past time it errored).
* `span.kind`: Always consumer kind.
* `error`: flag marking if it finished with error.

### Relevant Logs

The most relevan logs of the root span are the processing and error logs.
this is an example of a log in a processed object:

* `0.03ms: event=baggage, key=kubernetes.object.key, value=kube-system/kube-dns-7785f4d7dc-4brdg`
* `0.05ms: event=process_object, kubernetes.object.key=kube-system/kube-dns-7785f4d7dc-4brdg`
* `1.32s: event=error, message=randomly failed`
* `1.32s: event=forget, kubernetes.object.key=kube-system/kube-dns-7785f4d7dc-4brdg, message=max number of retries reached after failing, forgetting object key`
* `1.32s: success=false`

### Baggage

The [baggage][ot-baggage-url] is data that will be propagated in all the trace. Kooper adds one single baggage (it's expensive, must be careful), is the `kubernetes.object.key` that has the object key of the object being processed that started all the trace.

## Add/Delete handling span

After the root span, Kooper will create a new child span from this that will be a processing Add or a Delete depending of what kind of event is.

## Context on handler

Handler receives the context that has the parten span.

```go
type Handler interface {
    Add(context.Context, runtime.Object) error
    Delete(context.Context, string) error
}
```

You can get the parent span using opentracing API. For example:

```go
hand := &handler.HandlerFunc{
    AddFunc: func(ctx context.Context, obj runtime.Object) error {
        // Get the parent span.
        pSpan := opentracing.SpanFromContext(ctx)

        // Create a new span.
        span := tracer.StartSpan("AddFunc", opentracing.ChildOf(pSpan.Context()))
        defer span.Finish()

        // Do stuff...

        return nil
    },
    DeleteFunc: func(ctx context.Context, s string) error {
        // Get the parent span.
        pSpan := opentracing.SpanFromContext(ctx)

        // Create a new span.
        span := tracer.StartSpan("DeleteFunc", opentracing.ChildOf(pSpan.Context()))
        defer span.Finish()

        // Do stuff...

        return nil
    },
}
```

## Span diagram example

Kooper creates the first two spans and a user using kooper in it's controller creates 4 new spans.

```text
––|–––––––|–––––––|–––––––|–––––––|–––––––|–––––––|–––––––|–> time

 [ Kooper processJob ········································]
    [Kooper handleAddObject/handleDeleteObject···············]
      [User span1......]
                        [User span2·····]
                        [User span3··········]
                                              [User span4····]
```

## Tracer Service

The service name that identifies the traces created by kooper is managed by the tracer instance passed to the controller when creating the controller, this is that the tracer service name will be the trace service name.

## Example

Kooper comes with an [example][traced-controller], it's just the pod print controller example with faked calls to different services like Redis or AWS but tracing this calls. It uses the opentracing implementation of [Jaeger][jaeger-url].

First you need to run a development instance of Jaeger in localhost, the tracer passed to kooper controller will send the generated traces to this jaeger instance.

Run jaeger with docker.

```bash
docker run --rm -it \
    -p5775:5775/udp \
    -p6831:6831/udp \
    -p6832:6832/udp \
    -p5778:5778 \
    -p16686:16686 \
    -p14268:14268 \
    -p9411:9411 \
    jaegertracing/all-in-one:latest
```

Now just run the example, you need kubectl set to the cluster, it will make fake calls to faked services and will error randomly on the service calls, making retries on the object processing and creating different kind of traces.

```bash
go run ./examples/traced-controller/*
```

Check the created traces of the controller in [jaeger ui][jaeger-ui-local-url]

[opentracing-url][http://opentracing.io]
[ot-baggage-url][https://github.com/opentracing/specification/blob/master/specification.md#set-a-baggage-item]
[traced-controller][https://github.com/spotahome/kooper/tree/master/examples/traced-controller]
[jaeger-url][http://www.jaegertracing.io/]
[jaeger-ui-local-url][http://127.0.0.1:16686]