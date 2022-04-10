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

RUN curl -L -o /usr/bin/accrual \
    https://github.com/yandex-praktikum/go-musthave-diploma-tpl/raw/master/cmd/accrual/accrual_linux_amd64; \
    chmod +x /usr/bin/accrual

ENTRYPOINT ["/app/bin/server"]
