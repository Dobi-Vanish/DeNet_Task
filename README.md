Выполненное тестовое задание.  
Использованные технологии: chi, goose, jwt, bcrypt, godotenv, pgx.  
Для запуска проекта необходимо в косноли перейти в папку `project` и прописать `make up_build`. Всё запустится в Докере, миграции применятся автоматически путём использования `github.com/pressly/goose`.
DSN для подключения к БД и некоторые другие переменные берутся из .env файла.  
В проекте доступна функция аутентификации и регистрации, для доступа к ним не обязательно иметь JWT токен. 
Для доступа ко всем остальным путям необходимо токен. Он будет автоматически сгенерирован и отправлен при аутентификации пользователя в куках, перед этим необходимо зарегестрироваться.  
Успешная регистрация нового пользователя:  
  
![изображение](https://github.com/user-attachments/assets/8cb302b7-0050-43fa-ba58-2c878a021e24)  

Успешная аутентификация нового зарегистрированного пользователя:  

  ![изображение](https://github.com/user-attachments/assets/8439e6c6-499e-4212-a0dc-9218da87db48)  

`GET /users/{id}/status` Получение всей доступной информации об одном пользователе , id берётся из URL'a:    

  ![изображение](https://github.com/user-attachments/assets/1988bcac-23b3-42fc-be97-c9f222cfb5f5)  

  Добавим сразу второго пользователя:  

    ![изображение](https://github.com/user-attachments/assets/de433cb0-a050-4b5e-a6e8-c3ee3be15d1b)  

 `GET /users/leaderboard` - топ пользователей с самым большим балансом:  

   ![изображение](https://github.com/user-attachments/assets/d57a22db-e203-4a73-93a3-19e2d9b844e1)  

 `POST /users/{id}/task/complete` - выполнение какого-то задания, для примера представленно сразу несколько и их результат id берётся из URL'a:  
   ![изображение](https://github.com/user-attachments/assets/dcdbf492-694b-4db3-9205-9c11cacb9f11)  
   ![изображение](https://github.com/user-attachments/assets/1764cfb9-d875-4d6e-882c-929aaae4a697)  
   ![изображение](https://github.com/user-attachments/assets/73eef753-629f-4cb0-ac59-721473d95f12)   
   ![изображение](https://github.com/user-attachments/assets/22ecd4a6-ed91-41fe-b99f-154b60c5b41b)  
   
`POST /users/{id}/referrer` - ввод реферального кода. В задании не было чётко указано, как именно реализовать, поэтому у каждого пользователя есть реферальный код.
При вводе чьего-то реферального кода, тот, чей код введён, получает 100 очков, тот, кто вводил - получает 25 очков. Id берётся из URL'a, также проверяется чтобы пользователь не мог ввести для себя же свой же реферальный код:  
![изображение](https://github.com/user-attachments/assets/848ba847-cd9c-4778-ae66-3a1ecf68dbf7)  
![изображение](https://github.com/user-attachments/assets/d63f8b48-29f1-4e82-b3de-515c4c3634ae)  

  Также в коде присутсвуют и иные проверки на ввод данных, но упоминать их не стал.  







   

