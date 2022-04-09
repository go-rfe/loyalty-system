FROM golang:1.17

WORKDIR /app
COPY . /app/
RUN make build

RUN curl -L -o /usr/bin/statictest \
    https://github.com/Yandex-Practicum/go-autotests-bin/releases/latest/download/statictest; \
    chmod +x /usr/bin/statictest

RUN curl -L -o /usr/bin/gophermarttest \
    https://github.com/Yandex-Practicum/go-autotests-bin/releases/latest/download/gophermarttest; \
    chmod +x /usr/bin/gophermarttest

RUN curl -L -o /usr/bin/random \
    https://github.com/Yandex-Practicum/go-autotests-bin/releases/latest/download/random; \
    chmod +x /usr/bin/random

RUN curl -L -o /usr/bin/accural \
    https://github.com/yandex-praktikum/go-musthave-diploma-tpl/raw/master/cmd/accrual/accrual_linux_amd64; \
    chmod +x /usr/bin/accural

ENTRYPOINT ["go", "vet", "-vettool=/usr/bin/statictest", "./..."]
