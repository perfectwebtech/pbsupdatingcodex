use actix_web::{web, HttpResponse};
use uuid::Uuid;
use chrono::Utc;
use crate::{
    models::{CreateJobRequest, TranscodeJob, JobStatus, JobResponse},
    transcoder::TranscoderPool,
    Result,
};
use tracing::info;

pub async fn create_transcode_job(
    body: web::Json<CreateJobRequest>,
    pool: web::Data<TranscoderPool>,
    user_id: web::ReqData<i64>,
) -> Result<HttpResponse> {
    let user_id = *user_id.into_inner();

    let job = TranscodeJob {
        id: Uuid::new_v4(),
        user_id,
        input_file: body.input_file.clone(),
        output_file: None,
        preset: body.preset.clone(),
        status: JobStatus::Pending,
        progress: 0.0,
        error_message: None,
        created_at: Utc::now(),
        started_at: None,
        completed_at: None,
    };

    let job_id = job.id;
    pool.submit_job(job).await?;

    info!("Created transcode job: {} for user: {}", job_id, user_id);

    Ok(HttpResponse::Created().json(serde_json::json!({
        "success": true,
        "data": {
            "job_id": job_id,
            "status": "pending"
        }
    })))
}

pub async fn get_job(
    path: web::Path<Uuid>,
    pool: web::Data<TranscoderPool>,
) -> Result<HttpResponse> {
    let job_id = path.into_inner();

    if let Some(job) = pool.get_job(&job_id).await {
        Ok(HttpResponse::Ok().json(serde_json::json!({
            "success": true,
            "data": JobResponse {
                id: job.id,
                status: job.status,
                progress: job.progress,
                output_file: job.output_file,
                created_at: job.created_at,
            }
        })))
    } else {
        Ok(HttpResponse::NotFound().json(serde_json::json!({
            "success": false,
            "error": "Job not found"
        })))
    }
}

pub async fn get_job_status(
    path: web::Path<Uuid>,
    pool: web::Data<TranscoderPool>,
) -> Result<HttpResponse> {
    let job_id = path.into_inner();

    if let Some(job) = pool.get_job(&job_id).await {
        Ok(HttpResponse::Ok().json(serde_json::json!({
            "success": true,
            "data": {
                "job_id": job.id,
                "status": job.status,
                "progress": job.progress,
                "started_at": job.started_at,
                "completed_at": job.completed_at,
                "error": job.error_message,
            }
        })))
    } else {
        Ok(HttpResponse::NotFound().json(serde_json::json!({
            "success": false,
            "error": "Job not found"
        })))
    }
}

pub async fn cancel_job(
    path: web::Path<Uuid>,
    pool: web::Data<TranscoderPool>,
) -> Result<HttpResponse> {
    let job_id = path.into_inner();

    match pool.cancel_job(&job_id).await {
        Ok(_) => Ok(HttpResponse::Ok().json(serde_json::json!({
            "success": true,
            "message": "Job cancelled successfully"
        }))),
        Err(e) => Ok(HttpResponse::BadRequest().json(serde_json::json!({
            "success": false,
            "error": e.to_string()
        }))),
    }
}

pub async fn list_jobs() -> Result<HttpResponse> {
    // TODO: Implement job listing from database
    Ok(HttpResponse::Ok().json(serde_json::json!({
        "success": true,
        "data": {
            "jobs": [],
            "total": 0
        }
    })))
}

pub async fn metrics() -> HttpResponse {
    // TODO: Implement Prometheus metrics
    HttpResponse::Ok().body("# Transcoding Service Metrics\n")
}
