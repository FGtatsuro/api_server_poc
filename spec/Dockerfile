# FYI: https://future-architect.github.io/articles/20200513/
FROM python:3.9-slim-buster as builder

WORKDIR /opt/app

COPY requirements.lock /opt/app
RUN pip3 install -r requirements.lock


FROM python:3.9-slim-buster as runner

COPY --from=builder /usr/local/lib/python3.9/site-packages /usr/local/lib/python3.9/site-packages
RUN mkdir /spec
COPY test_api.py /spec

ENTRYPOINT ["python3", "-m", "pytest", "/spec"]
