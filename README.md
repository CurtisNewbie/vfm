# vfm

Virtual File Manager - A lightweight alternative of [file-service v1.2.7](https://github.com/CurtisNewbie/file-server/tree/v1.2.7). This app will run using schema originally created by `file-service v1.2.7`.

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

| Feature/Functionality                               | vfm                                      | file-service |
|-----------------------------------------------------|------------------------------------------|--------------|
| Manage File Storage                                 | not supported, `mini-fstore` is required | supported    |
| File Event Synchronization                          | not supported yet                        | supported    |
| File Package And Export (File Task)                 | not supported yet                        | supported    |
| Manage App Files (files that don't belong to users) | not supported                            | supported    |
| Virtual Folders (VFolders)                          | not supported yet                        | supported    |

## Migration From File-Service

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