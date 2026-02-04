package com.jobplatform.entity;

import java.time.LocalDateTime;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.PrePersist;
import jakarta.persistence.PreUpdate;
import jakarta.persistence.Table;
import lombok.Data;

/**
 * JPA Entity - представляет таблицу "jobs" в PostgreSQL.
 * 
 * Hibernate автоматически создаст таблицу с такими колонками:
 * - id (PRIMARY KEY, AUTO_INCREMENT)
 * - type (VARCHAR(50))
 * - status (VARCHAR(50))
 * - payload (TEXT)
 * - result (TEXT)
 * - error_message (TEXT)
 * - created_at (TIMESTAMP)
 * - updated_at (TIMESTAMP)
 */
@Data  // Lombok генерирует getters, setters, toString, equals, hashCode
@Entity
@Table(name = "jobs")
public class Job {
    
    /**
     * Primary Key. IDENTITY означает автоинкремент (PostgreSQL SERIAL).
     */
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    /**
     * Тип задачи: "HTTP_GET", "IMAGE_RESIZE", "SLEEP".
     */
    @Column(nullable = false, length = 50)
    private String type;
    
    /**
     * Текущий статус: "CREATED", "IN_PROGRESS", "COMPLETED", "FAILED".
     */
    @Column(nullable = false, length = 50)
    private String status = "CREATED";
    
    /**
     * JSON строка с параметрами задачи.
     * Пример: {"url": "https://example.com"} для HTTP_GET
     */
    @Column(nullable = false, columnDefinition = "TEXT")
    private String payload;
    
    /**
     * Результат выполнения от Go-воркера (приходит через gRPC).
     */
    @Column(columnDefinition = "TEXT")
    private String result;
    
    /**
     * Текст ошибки, если статус FAILED.
     */
    @Column(name = "error_message", columnDefinition = "TEXT")
    private String errorMessage;
    
    /**
     * Время создания записи.
     */
    @Column(name = "created_at", nullable = false, updatable = false)
    private LocalDateTime createdAt;
    
    /**
     * Время последнего обновления.
     */
    @Column(name = "updated_at", nullable = false)
    private LocalDateTime updatedAt;
    
    /**
     * Вызывается перед первым сохранением в БД.
     */
    @PrePersist
    protected void onCreate() {
        createdAt = LocalDateTime.now();
        updatedAt = LocalDateTime.now();
    }
    
    /**
     * Вызывается перед каждым обновлением записи.
     */
    @PreUpdate
    protected void onUpdate() {
        updatedAt = LocalDateTime.now();
    }
}
