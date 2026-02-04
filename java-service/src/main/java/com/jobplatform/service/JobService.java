package com.jobplatform.service;

import java.util.List;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional; // Обязательно!

import com.fasterxml.jackson.databind.ObjectMapper; // Обязательно!
import com.jobplatform.entity.Job;
import com.jobplatform.kafka.KafkaJobPublisher;
import com.jobplatform.repository.JobRepository;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

/**
 * Бизнес-логика для работы с задачами.
 */
@Slf4j  // Lombok создаст поле log для логирования
@Service
@RequiredArgsConstructor
public class JobService {
    
    private final JobRepository jobRepository;
    private final KafkaJobPublisher kafkaPublisher;
    
    // Jackson для конвертации payload объекта в JSON строку
    private final ObjectMapper objectMapper = new ObjectMapper();
    
    /**
     * Создаёт задачу и отправляет в Kafka.
     * 
     * Последовательность:
     * 1. Конвертируем payload (Map/Object) в JSON строку
     * 2. Сохраняем задачу в БД со статусом CREATED
     * 3. Отправляем Protobuf сообщение в Kafka топик
     * 4. Go-воркер получит это сообщение и начнёт обработку
     * 
     * @Transactional гарантирует, что запись в БД откатится при ошибке Kafka.
     */
    @Transactional
    public Job createJob(String type, Object payload) {
        try {
            // Преобразуем Map или Object в JSON строку
            // Пример: {"duration_ms": 3000} -> "{\"duration_ms\":3000}"
            String payloadJson = objectMapper.writeValueAsString(payload);
            
            // Создаём entity
            Job job = new Job();
            job.setType(type.toUpperCase());  // Нормализуем к верхнему регистру
            job.setPayload(payloadJson);
            job.setStatus("CREATED");
            
            // Сохраняем в PostgreSQL (получим сгенерированный ID)
            job = jobRepository.save(job);
            log.info("Created job {} with type {}", job.getId(), type);
            
            // Отправляем в Kafka (Go-воркер получит и обработает)
            kafkaPublisher.publishJob(job);
            
            return job;
            
        } catch (Exception e) {
            log.error("Failed to create job", e);
            throw new RuntimeException("Failed to create job: " + e.getMessage());
        }
    }
    
    /**
     * Получение задачи по ID.
     * 
     * Бросает exception, если задача не найдена (Spring превратит в HTTP 500).
     * В реальном проекте лучше использовать кастомный JobNotFoundException -> 404.
     */
    public Job getJob(Long id) {
        return jobRepository.findById(id)
            .orElseThrow(() -> new RuntimeException("Job not found: " + id));
    }
    
    /**
     * Список всех задач.
     */
    public List<Job> getAllJobs() {
        return jobRepository.findAll();
    }
    
    /**
     * Обновление статуса задачи.
     * 
     * Вызывается из gRPC сервиса, когда Go-воркер завершил обработку.
     * 
     * @param jobId ID задачи
     * @param status "COMPLETED" или "FAILED"
     * @param result Результат выполнения (может быть null)
     * @param errorMessage Текст ошибки (только для FAILED)
     */
    @Transactional
    public void updateJobStatus(Long jobId, String status, String result, String errorMessage) {
        Job job = jobRepository.findById(jobId)
            .orElseThrow(() -> new RuntimeException("Job not found: " + jobId));
        
        // Обновляем поля
        job.setStatus(status);
        job.setResult(result);
        job.setErrorMessage(errorMessage);
        
        // Hibernate автоматически сохранит изменения в БД при завершении транзакции
        jobRepository.save(job);
        log.info("Updated job {} to status {}", jobId, status);
    }
}
