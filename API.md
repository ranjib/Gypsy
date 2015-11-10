## REST end points
 This document enlist the REST endpoints served by the gypsy http
 server. It is used by the dahsboatd ui, dashboard ui, and clients, to
 manupulate pipeline configurations, artifacts, build runs etc.



### Manage pipelines

-	GET /pipelines
  List pipelines (json format)

-	GET /pipelines/{pipeline_name}
  Get a pipeline configurations (yaml format)

-	PUT /pipelines
  Create a pipeline from yaml configuration specification

-	DELETE /pipelines/{pipeline_name}
  Delete a pipeline

-	POST /pipelines/{pipeline_name}
  Update a pipeline configuration (yaml format)

### Manage pipeline runs

-	GET /pipelines/{pipeline_name}/runs
  List all build runs for a pipeline

-	GET /pipelines/{pipeline_name}/runs/{run_id}
  Get run details of a pipeline build

-	POST /pipelines/{pipeline_name}/runs/{run_id}
  Create run details for a pipeline build (used by build agents)

-	DELETE /pipelines/{pipeline_name}/runs/{run_id}
  Delete a build run for a given pipeline
	
### Manage pipeline run artifacts

-	GET /pipelines/{pipeline_name}/runs/{run_id}/artifacts
  List artifacts for a particular pipeline run

-	GET /pipelines/{pipeline_name}/runs/{run_id}/artifacts/{artifact_name}
  Get a particular artifact from a pipeline run

-	POST /pipelines/{pipeline_name}/runs/{run_id}/artifacts/{artifact_name}
  Upload artifact for a particular pipeline run

-	DELETE /pipelines/{pipeline_name}/runs/{run_id}/artifacts/{artifact_name}
  Delete artifact of a particular pipeline run
