### Setup

Fork the project into your personal repository. Clone it to local computer.

```sh
git clone https://github.com/etifontaine/istio-redirector.git

cd istio-redirector/front
npm run install

cd istio-redirector/server
go mod download
```

### Project configuration

#### Front side

- Create a `.env.local` file with `NEXT_PUBLIC_API_BASE_URL="http://127.0.0.1:8080/"`.

#### Server side

- Create your configuration, `mv config_example.yaml config.yaml`, and edit the value following your setup.