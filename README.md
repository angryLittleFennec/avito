# Avito Internship 2024

## Описание проекта

Этот проект представляет собой тестовое задание для отбора на стажировку в Авито,  заключающееся в имплементации API-сервиса, предназначенного для управления баннерами и их связью с тегами и фичами. Сервис позволяет создавать, обновлять, удалять и получать информацию о баннерах, а также управлять их активностью. В рамках данной задачи были выполненны все основные пункты технического задания (1-6), а также был добавлен линтер и реализованы e2e тесты, покрывающие все ручки (п. 5, 6 из дополнительных заданий)

## Как запустить

Для локального запуска и тестирования проекта используется `Docker` и `Docker Compose`. Примеры команд:

```bash
make run    # Запуск сервисов
make test   # Запуск тестов
```

## Архитектура

### Сервер

Серверная часть реализована на языке `Go` с использованием фреймворка `Echo`. Она включает в себя обработку HTTP-запросов и взаимодействие с базой данных `PostgreSQL`. Для хранения временных данных используется `Redis`. Интерфейс методов сервера и типы получаемых данных сгенерированны с помощью `oapi-codegen`. Интерфейс ручек был реализован в соответствии с техническим заданием.

### Авторизация

Для авторизации в нашем API используются предопределённые токены: `admin1` для администраторских действий и `user1` для пользовательских. Авторизация выполняется с помощью middleware, который проверяет наличие и корректность токена в заголовке запроса.

### База Данных

Для работы с `PostgreSQL` базой данных использовался `gorm`, были созданы две модели Banner для баннеров и BannerFeatureTag для связи баннера с тегами и фичами. На вторую модель наложено такое ограничение, что пары фича-тег не могут повторяться при помощи unique index. В случае ошибки в POST или PATCH запросе, вызванной данным ограничением, мы возвращаем код ошибки 409 статус Conflict. Миграции происходят автоматически при помощи `gorm`

## CI/CD

В `CI GitHub Actions` реализована проверка линтера и запуск e2e тестов.

### Workflow конфигурации

- **Lint**: Запускается на каждое изменение в репозитории для проверки стиля кода.
- **e2e-tests**: Выполняет e2e тесты после каждого изменения для проверки корректности работы API.

## Тестирование

Для тестирования используются модули `testing`, `testify`. Клиент для тестов был также сгенерирован при помощи `oapi-codegen` из данного API файла. Тесты включают проверки функций API на соответствие ожидаемому поведению. Реализованы различные end-to-end (e2e) тесты, которые покрывают функциональные аспекты работы с баннерами, включая создание, получение, обновление и удаление баннеров.

## Описание тестов

- ### TestBannerLifecycle

    Тест проверяет возможность получения баннера после создания.

- ### TestDeleteBannerLifecycle

    Тест проверяет процесс создания и последующего удаления баннера. После удаления баннер отсутствует среди получаемых баннеров.

- ### TestPatchAndUpdateBanner

    Тест на обновление данных баннера (изменение содержимого, флага активности и списка тегов)

- ### TestGetUserBanner

    Тест на получение баннера пользователем.

- ### TestPostDuplicateBanner

    Тест на проверку обработки создания дубликатов баннеров (ожидается получение статуса 409 (Conflict), указывающего на нарушение уникальности данных).


## Запуск тестов

Тесты запускаются с использованием команды `make test`, которая определена в файле `Makefile` и выполняет команду `go test`, запуская все тесты в проекте. После завершения тестов выполняется команда `docker compose down --volumes` для остановки и очистки всех используемых сервисов и данных.
