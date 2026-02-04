package com.jobplatform.kafka;

import com.jobplatform.entity.Job;
// ИСПРАВЛЕНИЕ: Импортируем сгенерированные классы напрямую
import com.jobplatform.grpc.JobTask;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Component;

import java.time.ZoneOffset;

@Slf4j
@Component
@RequiredArgsConstructor
public class KafkaJobPublisher {
    
    private final KafkaTemplate<String, byte[]> kafkaTemplate;
    
    @Value("${kafka.topic}")
    private String topic;
    
    public void publishJob(Job job) {
        try {
            // Используем внутренний Enum из JobTask
            JobTask.TaskType taskType = mapType(job.getType());
            
            JobTask message = JobTask.newBuilder()
                .setJobId(job.getId())
                .setType(taskType)
                .setPayload(job.getPayload())
                .setCreatedAt(job.getCreatedAt().toEpochSecond(ZoneOffset.UTC))
                .build();
            
            kafkaTemplate.send(topic, String.valueOf(job.getId()), message.toByteArray());
            log.info("Published job {} to Kafka", job.getId());
            
        } catch (Exception e) {
            log.error("Failed to publish job {} to Kafka", job.getId(), e);
            throw new RuntimeException("Kafka publish failed", e);
        }
    }
    
    private JobTask.TaskType mapType(String type) {
        return switch (type) {
            case "HTTP_GET" -> JobTask.TaskType.HTTP_GET;
            case "IMAGE_RESIZE" -> JobTask.TaskType.IMAGE_RESIZE;
            case "SLEEP" -> JobTask.TaskType.SLEEP;
            default -> JobTask.TaskType.UNKNOWN_TYPE;
        };
    }
}
