package com.jobplatform.controller;

import com.jobplatform.entity.Job;
import com.jobplatform.service.JobService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;
/**
 * REST API контроллер для управления задачами.
 * 
 * Доступные endpoints:
 * - POST /api/v1/jobs - Создать новую задачу
 * - GET /api/v1/jobs/{id} - Получить задачу по ID
 * - GET /api/v1/jobs - Список всех задач
 */
@RestController
@RequestMapping("/api/v1/jobs")
@RequiredArgsConstructor  // Lombok создаст конструктор с final полями (для DI)
public class JobController {
    
    private final JobService jobService;
    
    /**
     * Создание новой задачи.
     * 
     * Ожидает JSON:
     * {
     *   "type": "SLEEP",
     *   "payload": {
     *     "duration_ms": 3000
     *   }
     * }
     * 
     * Возвращает HTTP 202 Accepted (задача принята к обработке, но ещё не выполнена).
     */
    @PostMapping
    public ResponseEntity<Job> createJob(@RequestBody Map<String, Object> request) {
        String type = (String) request.get("type");
        Object payload = request.get("payload");  // Может быть Map, List или примитив
        
        Job job = jobService.createJob(type, payload);
        
        // 202 Accepted - стандартный код для асинхронных операций
        return ResponseEntity.status(HttpStatus.ACCEPTED).body(job);
    }
    
    /**
     * Получение задачи по ID.
     * 
     * Пример: GET /api/v1/jobs/42
     * 
     * Вернёт 404, если задача не найдена.
     */
    @GetMapping("/{id}")
    public ResponseEntity<Job> getJob(@PathVariable Long id) {
        Job job = jobService.getJob(id);
        return ResponseEntity.ok(job);
    }
    
    /**
     * Список всех задач в системе.
     * 
     * Пример: GET /api/v1/jobs
     * 
     * В реальном проекте стоит добавить пагинацию (Pageable).
     */
    @GetMapping
    public ResponseEntity<List<Job>> getAllJobs() {
        return ResponseEntity.ok(jobService.getAllJobs());
    }
}
