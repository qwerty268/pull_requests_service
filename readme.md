Только создаю новую команду (НЕ ОБНОВЛЯЮ)
т к в ответах указали ошибку с ответом: команда уже существует
```yaml
paths:
  /team/add: 
    '400':
      description: Команда уже существует
      content:
        application/json:
          schema: 
            $ref: '#/components/schemas/ErrorResponse'
          example:
            error:
              code: TEAM_EXISTS
              message: team_name already exists
