# vfm

Virtual File Manager - A lightweight alternative of [file-service v1.2.7](https://github.com/CurtisNewbie/file-server).

Unlike file-service, vfm doesn't manage the actual file storage. The file storage is managed by [mini-fstore](https://github.com/CurtisNewbie/mini-fstore), a light-weight and simple solution designed for general usage.

## Requirements

<!-- TODO upgrade file-service-front to v1.2.3, the v1.2.3 is not implemented yet -->

- file-service-front (Angular frontend) >= [v1.2.2](https://github.com/CurtisNewbie/file-service-front/tree/v1.2.2)
- auth-gateway >= [v1.1.1](https://github.com/CurtisNewbie/auth-gateway/tree/v1.1.1)
- auth-service >= [v1.1.6](https://github.com/CurtisNewbie/auth-service/tree/v1.1.6)
- goauth >= [v1.0.0](https://github.com/CurtisNewbie/goauth/tree/v1.0.0)
- mini-fstore(https://github.com/CurtisNewbie/mini-fstore)
- MySQL 5.7 or 8
- Consul
- RabbitMQ
- Redis

## Configuration

| Property | Description | Default Value |
|----------|-------------|---------------|
|          |             |               |

## Difference between vfm and file-service

| Feature/Functionality      | vfm                                    | file-service |
|----------------------------|----------------------------------------|--------------|
| Manage File Storage        | not supported, mini-fstore is required | supported    |
| File Event Synchronization | not supported yet                      | supported    |
| File Package And Export    | not supported yet                      | supported    |