# vfm

Virtual File Manager - A lightweight replacement for [file-service v1.2.7](https://github.com/CurtisNewbie/file-server/tree/v1.2.7). This app will run using the schema originally created by `file-service v1.2.7`.

Unlike file-service, vfm doesn't manage the actual file storage. The file storage is managed by [mini-fstore](https://github.com/CurtisNewbie/mini-fstore), a light-weight and simple solution designed for general usage.

See [CurtisNewbie/system-overview](https://github.com/CurtisNewbie/system-overview) for the big picture.

## Requirements

- [user-vault](https://github.com/CurtisNewbie/user-vault)
- [goauth](https://github.com/CurtisNewbie/goauth)
- [mini-fstore](https://github.com/CurtisNewbie/mini-fstore)
- [hammer](https://github.com/CurtisNewbie/hammer)
- MySQL
- Consul
- Redis
- RabbitMQ

## Configuration

Check [miso](https://github.com/curtisnewbie/miso) for more about configuration.

| Property | Description | Default Value |
|----------|-------------|---------------|
|          |             |               |

## Thumbnail Generation

Whenever a file is uploaded to `mini-fstore`, and a file record is created on `vfm`, `vfm` checks whether it's potentially an image (by the file name). If so, it sends a MQ message to `hammer` (via RabbitMQ) for image compression. The generated thumbnail is uploaded to `mini-fstore` by `hammer`, and a MQ message replied by `hammer` tells what the file_id (on mini-fstore) of the thumbnail is.

To compensate the thumbnail generation for historical records, uses following curl command to trigger the compensation process.

```sh
curl -X POST "http://localhost:8086/compensate/image/compression"
```

## Updates

- Since v0.0.4, `vfm` relies on `evnet-pump` to listen to the binlog events. Whenever a new `file_info` record is inserted, the `event-pump` sends MQ to `vfm`, which triggers the image compression workflow if the file is an image.
- Since v0.0.8
    - Users can only share files using `vfolder`, field `file_info.user_group` and table `file_sharing` are deprecated.
    - Uploaders's user no can be updated (compensated) using: `curl -X POST "http://localhost:8086/compensate/file/uploaderno"`.

## Doc

- [File-Service Migration](./migration/file-service-migration.md)
- [Fantahsea Migration](./migration/fantahsea-migration.md)