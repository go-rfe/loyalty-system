FROM golang:1.17

WORKDIR /

RUN curl -L -o /usr/bin/gophermarttest \
    https://github.com/Yandex-Practicum/go-autotests-bin/releases/latest/download/gophermarttest; \
    chmod +x /usr/bin/gophermarttest

RUN curl -L -o /usr/bin/accrual \
    https://github.com/yandex-praktikum/go-musthave-diploma-tpl/raw/master/cmd/accrual/accrual_linux_amd64; \
    chmod +x /usr/bin/accrual
