# vfm

vfm stands for virtual file manager. vfm doesn't manage the actual file storage, the actual files are stored by [mini-fstore](https://github.com/CurtisNewbie/mini-fstore).

## Requirements

- [event-pump](https://github.com/CurtisNewbie/event-pump)
- [user-vault](https://github.com/CurtisNewbie/user-vault)
- [mini-fstore](https://github.com/CurtisNewbie/mini-fstore)
- [hammer](https://github.com/CurtisNewbie/hammer)
- MySQL
- Consul
- Redis
- RabbitMQ

## Configuration

Check [miso](https://github.com/curtisnewbie/miso) and [gocommon](https://github.com/curtisnewbie/gocommon) for more about configuration.

| Property | Description | Default Value |
| -------- | ----------- | ------------- |
|          |             |               |


## Updates

- Since v0.0.4, `vfm` relies on `evnet-pump` to listen to the binlog events. Whenever a new `file_info` record is inserted, the `event-pump` sends MQ to `vfm`, which triggers the image compression workflow if the file is an image.
- Since v0.0.8
    - Users can only share files using `vfolder`, field `file_info.user_group` and table `file_sharing` are deprecated.
- Since v0.1.3, [fantahsea](https://github.com/curtisnewbie/fantahsea) has been merged into vfm codebase, see [Fantahsea Migration](./doc/fantahsea-migration.md).
- Since v0.1.17, file tag functionality is removed.

## Maintenance

Calculate size of all directories recursively, bubbling up to the root:

```sh
curl -X POST "http://localhost:8086/compensate/dir/calculate-size"
```

Compensate thumbnail generations, those that are images/videos (guessed by names) are processed to generate thumbnails:

```sh
curl -X POST "http://localhost:8086/compensate/thumbnail"
```

## Schema Migration

Everytime the schema is changed, a new SQL script for that specific version is maintained at `internal/schema/scripts`. The migration is automatically handled by [github.com/curtisnewbie/svc](https://github.com/curtisnewbie/svc).

## Doc

- [File-Service Migration](./doc/file-service-migration.md)
- [Fantahsea Migration](./doc/fantahsea-migration.md)
- [Thumbnail Generation](./doc/thumbnail.md)
