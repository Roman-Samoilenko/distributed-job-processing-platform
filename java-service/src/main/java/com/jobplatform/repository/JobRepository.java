package com.jobplatform.repository;

import com.jobplatform.entity.Job;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

/**
 * Репозиторий для работы с таблицей jobs.
 * 
 * Spring Data JPA автоматически создаёт реализацию с методами:
 * - save(job) - INSERT или UPDATE
 * - findById(id) - SELECT WHERE id = ?
 * - findAll() - SELECT * FROM jobs
 * - deleteById(id) - DELETE WHERE id = ?
 * 
 * Можно добавлять кастомные методы, Spring сгенерирует SQL по имени метода:
 * - findByStatus(String status) -> SELECT * WHERE status = ?
 * - findByTypeAndStatus(String type, String status) -> SELECT * WHERE type = ? AND status = ?
 */
@Repository
public interface JobRepository extends JpaRepository<Job, Long> {
    // Пока используем только стандартные методы
}
