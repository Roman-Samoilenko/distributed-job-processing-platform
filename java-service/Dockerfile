# Этап 1: Сборка
FROM gradle:8.5-jdk17 AS build
WORKDIR /app

# Создаем структуру папок заранее
RUN mkdir -p java-service proto

# 1. Копируем файлы сборки (путь относительно корня проекта)
COPY java-service/build.gradle java-service/settings.gradle ./java-service/

# 2. Копируем ОБЩУЮ папку proto
COPY proto ./proto/

# 3. Копируем исходный код
COPY java-service/src ./java-service/src/

# 4. Собираем проект
WORKDIR /app/java-service
RUN gradle build --no-daemon -x test

# Этап 2: Запуск
FROM eclipse-temurin:17-jre-alpine
WORKDIR /app
COPY --from=build /app/java-service/build/libs/*.jar app.jar

EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]
