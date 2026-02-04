package com.jobplatform;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

/**
 * Точка входа в Spring Boot приложение.
 * 
 * @SpringBootApplication включает:
 * - @Configuration: этот класс может содержать Spring beans
 * - @EnableAutoConfiguration: автоматически настраивает Spring на основе зависимостей
 * - @ComponentScan: ищет компоненты (@Controller, @Service, @Repository) в этом пакете
 */
@SpringBootApplication
public class JobPlatformApplication {
    
    public static void main(String[] args) {
        // Запускает Spring контейнер и встроенный Tomcat сервер
        SpringApplication.run(JobPlatformApplication.class, args);
    }
}
