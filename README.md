# Git Connector

GitHub-GitLab-Connector нужен для того, чтобы вызывать Gitlab CI в GitHub.

## Настройка GitHub App и Webhooks

Чтобы использовать GitHub-GitLab-Connector, для начала необходимо создать GitHub App. GitHub-GitLab-Connector - пока поддерживает обслуживание только одного репозитория. Поэтому для каждого нового репозитория необходимо будет создавать новую GitHub App и поднимать новый GitHub-GitLab-Connector с новым URL. 

При создании приложения указываем `Homepage URL` и `User authorization callback URL` указываем полный URL до нашего GitHub-GitLab-Connector. `Webhook URL` - указываем полный URL + `/githubhooks`. Создаём `secret` для хуков и запоминаем его. Это URL, куда будут отправлятся хуки с GitHub. Далее в Permissions даём разрешение на чтение и запись для Checks API и Contents API. И ставим галочки на получение событий об `Check Suite` и `Push`.

Сгенерировать и скачать PEM файл с приватным ключом для GitHub приложения.

Устанавливаем GitHub App для нужного репозитория.

Так же необходимо настроить webhook на GitLab репозитории, куда будет зеркалироваться GutHub репозиторий. В `Webhook URL` прописываем тот же полный урл + `/gitlabhooks`. Необходимо отправлять только `PipeLineEvents`. Создаём `secret` для хуков, такой же как и для GitHub.

## GitHub-GitLab-Connector build

```bash
git clone https://gitlab.pixelplex.by/service/github-gitlab-connector.git

cd github-gitlab-connector

# Enable build from any dir (not only GOPATH)
export GO111MODULE=on 

# With that cmd all deps will be downloaded
go build main.go params.go

# Check work
./main --help

```

## Запуск GitHub-GitLab-Connector

Первым делом для корректной работы необходимо добавить ssh ключ сервера, где будет запускаться GitHub-GitLab-Connector, в GitHub и GitLab.

Параметры передаваемое командной строкой:

* `--github` - GitHub URL, с которого будет происходить зеркалирование. URL должен быть SSH-формата.
* `--gitlab` - GitHub URL, с которого будет происходить зеркалирование. URL должен быть SSH-формата.
* `--local-path` - локальный путь на сервере, куда будет клонироваться репозиторий.
* `--port` - порт, который будет слушать GitHub-GitLab-Connector.
* `--privkey` - путь к PEM файлу, в котором лежит приватный ключ, сгенерированный при создании GitHub App.
* `--secret` - secret, который использовался для создания webhooks.

Посмотреть стандартные значения параметров можно `./main --help`.

Сервис должен запуститься без всяких ошибок.

Когда происходит push в GitHub репозиторий - репозиторий на GitLab обновится автоматически и запустит gitlab-ci, если подливался новый коммит. Если в GitHub подливается новый коммит - то по событию Check_suite создастся check_run, который будет находиться в `in_progress`, пока не получит event о изменении статуса pipline с GitLab.