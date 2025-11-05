use std::sync::Arc;
use tokio::sync::{Semaphore, RwLock};
use uuid::Uuid;
use std::collections::HashMap;
use tracing::{info, error, warn};
use crate::models::{TranscodeJob, JobStatus};
use crate::Result;

pub struct TranscoderPool {
    semaphore: Arc<Semaphore>,
    jobs: Arc<RwLock<HashMap<Uuid, Arc<RwLock<TranscodeJob>>>>>,
}

impl TranscoderPool {
    pub fn new(max_concurrent: usize) -> Self {
        Self {
            semaphore: Arc::new(Semaphore::new(max_concurrent)),
            jobs: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    pub async fn submit_job(&self, job: TranscodeJob) -> Result<()> {
        let job_id = job.id;
        let job_arc = Arc::new(RwLock::new(job));

        {
            let mut jobs = self.jobs.write().await;
            jobs.insert(job_id, job_arc.clone());
        }

        let semaphore = self.semaphore.clone();
        let jobs_map = self.jobs.clone();

        tokio::spawn(async move {
            let _permit = semaphore.acquire().await.unwrap();

            info!("Starting transcode job: {}", job_id);

            // Update job status to processing
            {
                let mut job = job_arc.write().await;
                job.status = JobStatus::Processing;
                job.started_at = Some(chrono::Utc::now());
            }

            // Perform transcoding
            match Self::transcode(job_arc.clone()).await {
                Ok(_) => {
                    info!("Transcode job completed: {}", job_id);
                    let mut job = job_arc.write().await;
                    job.status = JobStatus::Completed;
                    job.progress = 100.0;
                    job.completed_at = Some(chrono::Utc::now());
                }
                Err(e) => {
                    error!("Transcode job failed: {} - {:?}", job_id, e);
                    let mut job = job_arc.write().await;
                    job.status = JobStatus::Failed;
                    job.error_message = Some(e.to_string());
                    job.completed_at = Some(chrono::Utc::now());
                }
            }

            // Clean up after some time
            tokio::time::sleep(tokio::time::Duration::from_secs(3600)).await;
            let mut jobs = jobs_map.write().await;
            jobs.remove(&job_id);
        });

        Ok(())
    }

    async fn transcode(job: Arc<RwLock<TranscodeJob>>) -> Result<()> {
        // Simulate transcoding with progress updates
        // In production, this would use FFmpeg bindings

        let job_read = job.read().await;
        let input_file = job_read.input_file.clone();
        let preset = job_read.preset.clone();
        drop(job_read);

        info!("Transcoding {} with preset {}", input_file, preset);

        // Simulate progressive transcoding
        for i in 0..10 {
            tokio::time::sleep(tokio::time::Duration::from_secs(2)).await;

            let mut job_write = job.write().await;
            job_write.progress = (i + 1) as f32 * 10.0;

            info!("Job {} progress: {}%", job_write.id, job_write.progress);
        }

        // Set output file
        let mut job_write = job.write().await;
        job_write.output_file = Some(format!("output_{}.mp4", job_write.id));

        Ok(())
    }

    pub async fn get_job(&self, job_id: &Uuid) -> Option<TranscodeJob> {
        let jobs = self.jobs.read().await;
        if let Some(job_arc) = jobs.get(job_id) {
            let job = job_arc.read().await;
            Some(job.clone())
        } else {
            None
        }
    }

    pub async fn cancel_job(&self, job_id: &Uuid) -> Result<()> {
        let jobs = self.jobs.read().await;
        if let Some(job_arc) = jobs.get(job_id) {
            let mut job = job_arc.write().await;
            if job.status == JobStatus::Processing || job.status == JobStatus::Pending {
                job.status = JobStatus::Cancelled;
                job.completed_at = Some(chrono::Utc::now());
                info!("Job {} cancelled", job_id);
                Ok(())
            } else {
                Err(crate::Error::InvalidJobState(
                    "Job cannot be cancelled in current state".to_string()
                ))
            }
        } else {
            Err(crate::Error::JobNotFound)
        }
    }
}
