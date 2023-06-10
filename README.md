# vfm

Virtual File Manager - A lightweight replacement for [file-service v1.2.7](https://github.com/CurtisNewbie/file-server/tree/v1.2.7). This app will run using the schema originally created by `file-service v1.2.7`.

Unlike file-service, vfm doesn't manage the actual file storage. The file storage is managed by [mini-fstore](https://github.com/CurtisNewbie/mini-fstore), a light-weight and simple solution designed for general usage.

## Requirements

- [bolobao (Angular Frontend)](https://github.com/curtisnewbie/bolobao)
- [auth-gateway >= v1.1.1](https://github.com/CurtisNewbie/auth-gateway/tree/v1.1.1)
- [auth-service >= v1.1.6](https://github.com/CurtisNewbie/auth-service/tree/v1.1.6)
- [goauth >= v1.0.0](https://github.com/CurtisNewbie/goauth/tree/v1.0.0)
- [mini-fstore >= v0.0.2](https://github.com/CurtisNewbie/mini-fstore/tree/v0.0.2)
- [hammer >= v0.0.1](https://github.com/CurtisNewbie/hammer)
- MySQL 5.7 or 8
- Consul
- Redis

## Configuration

Check [gocommon](https://github.com/curtisnewbie/gocommon) for more about configuration.

| Property | Description | Default Value |
|----------|-------------|---------------|
|          |             |               |

## Difference between vfm and file-service

| Feature/Functionality                               | vfm                                                       | file-service |
|-----------------------------------------------------|-----------------------------------------------------------|--------------|
| Manage User Files (virtually)                       | supported                                                 | supported    |
| Virtual Folders (VFolders)                          | supported                                                 | supported    |
| Virtual Directories                                 | supported                                                 | supported    |
| Manage File Storage                                 | not supported, `mini-fstore` is required                  | supported    |
| Manage App Files (files that don't belong to users) | not supported, should be handled by `mini-fstore` instead | supported    |
| File Event Synchronization                          | not supported, should be handled by `mini-fstore` instead | supported    |
| File Package And Export (File Task)                 | not supported                                             | supported    |

## Migration from File-Service

If `vfm` is migrated from `file-service`, the following SQL script should be executed:

```sql
ALTER TABLE file_info ADD COLUMN fstore_file_id VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'mini-fstore file id';
```

Then in `mini-fstore`, add the following configuration:

```yaml
fstore:
  migr:
    file-server:
      storage: ${PATH_TO_FILE_SERVER_FILES}
      enabled: true  # enable migration
      dry-run: false # disable dry-run
      mysql:
        user: ${USERNAME}
        password: ${PASSWORD}
        database: ${FILE_SERVER_DATABASE_NAME}
        host: ${HOST}
        port: ${PORT}
```

For more information, please read mini-fstore's [README](https://github.com/CurtisNewbie/mini-fstore).

## Thumbnail Generation

Whenever a file is uploaded to `mini-fstore`, and a file record is created on `vfm`, `vfm` checks whether it's potentially an image (by the file name). If so, it sends a MQ message to `hammer` (via RabbitMQ) for image compression. The generated thumbnail is uploaded to `mini-fstore` by `hammer`, and a MQ message replied by `hammer` tells what the file_id (on mini-fstore) of the thumbnail is.

To compensate the thumbnail generation for historical records, uses following curl command to trigger the compensation process.

```sh
curl -X POST "http://localhost:8086/compensate/image/compression"
```