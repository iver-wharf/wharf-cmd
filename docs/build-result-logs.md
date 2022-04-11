# Build result logs

Logs are retrieved by wharf-cmd-worker and then saved to disk so that other
components, such as wharf-cmd-aggregator, can download them safely when they
feel ready.

If you run `wharf` locally then you can access the build logs safely in your
editor of choice after the build is completed. The location of these logs depend
on your OS (Windows vs GNU/Linux), but is best sought after in wharf-cmd's help
text by running `wharf run --help`.

## Streaming

The logs are streamed from the Kubernetes pods or Docker containers (depending=
on the execution environment of your choice) and gets piped into two locations:

1. A file on disk.
2. To all open channels (e.g gRPC connections that are listening to logs).

Whenever a new client connects to the wharf-cmd-worker to stream its logs it
will start by streaming from the files on disk to let the client catch up on all
logs since the start of the build.

This will cause logs to be received out of order if the build has not yet
completed, as the streaming of old logs is performed in parallel to new logs
being received.

Consider the following diagram, where the wharf-cmd-aggregator streams logs from
wharf-cmd-worker via gRPC, which in turn is streaming the logs from the
container via Kubernetes API or Docker socket.

```text
  (aggregator)        (worker)          (container)
       |                  |                  |
       |                  |-- stream logs -->|
       |                  |<-- log line 1 ---|
       |                  |<-- log line 2 ---|
       |                  |<-- log line 3 ---|
       |-- stream logs -->|                  |
       |<-- log line 1 ---|                  |
       |                  |<-- log line 4 ---|
       |<-- log line 4 ---|                  |
       |<-- log line 2 ---|                  |
       |                  |<-- log line 5 ---|
       |<-- log line 5 ---|                  |
       |<-- log line 3 ---|                  |
       |             [caught up]             |
       |                  |<-- log line 6 ---|
       |<-- log line 6 ---|                  |
       |                  |<-- log line 7 ---|
       |<-- log line 7 ---|                  |
       V                  V                  V
```

Be weary of this when streaming the logs from the wharf-cmd-worker. If you are
displaying the log output from the wharf-cmd-worker directly then it's a good
idea to keep your own buffer and wait for the logs to catch up before starting
to display the logs.

## Why not always proxy?

The reasons for doing this instead of always proxying all logs from the
container (in Kubernetes pod or local Docker container) for all new streams
from the aggregator is:

- Buffers the logs locally to rely on the disk IO instead of the network IO as
  it's more reliable.

- Allows access to logs after running build locally without the
  wharf-cmd-aggregator.

- Kubernetes and Docker uses log rotation to discard old logs. As we are using
  the log line index+1 as the log ID, such log rotation would ruin the
  consistency in the logs in the database.
