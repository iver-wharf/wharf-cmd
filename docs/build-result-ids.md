# Build result IDs

The wharf-cmd-worker provides IDs for logs, status updates, and artifacts. These
IDs are unique on a step-basis and will cause collisions if they are used to
index build results from multiple build steps.

## Definitions

- Log ID: Line number in the log file for the step.
- Status update ID: Index+1 of the status update.
- Artifact ID: Index+1 of the status update.

"Index+1" means the value is incremented for each added status update or
artifact for that step, with the first value being 1.

## Unique IDs

To get a unique ID of the build results, each result ID has to be the union of:

- build ID (as provided by wharf-api)
- step ID (as provided by wharf-cmd-worker)
- result ID (log ID, status update ID, artifact ID)

If storing this in the database with a "unique constraint", either concatenate
the fields as `buildID-stepID-resultID` (e.g `123-1-531`) or use a composite
index ([GORM composite indexes docs](https://gorm.io/docs/indexes.html#Composite-Indexes))
to guarantee no collisions.

Build results of different types for the same step may have colliding IDs, so
if storing all results in the same table then the type is also needed, ex:

- `123-1-log-3`
- `123-1-status-3`
- `123-1-artifact-3`
