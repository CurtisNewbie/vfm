# File-service migration

vfm was originally developed as a lightweight replacement for [file-service v1.2.7](https://github.com/CurtisNewbie/file-server/tree/v1.2.7). It runs using the schema originally created by `file-service v1.2.7` (of course the schema changed in the following releases).

Unlike file-service, vfm doesn't manage the actual file storage. The file storage is managed by [mini-fstore](https://github.com/CurtisNewbie/mini-fstore), a light-weight and simple solution designed for general usage.


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