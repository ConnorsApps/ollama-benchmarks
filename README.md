# ollama-benchmarks
This repository is dedicated to testing prompts against different [Ollama](https://ollama.com/) models.

Key Features:
- Results of the tests are uploaded to a [Minio](https://min.io/) bucket
- The tests are executed within [Kubernetes jobs](./k8s-job.yaml)
- The program is written in [Golang](https://go.dev/)
