alter table file_info drop column uploader_id,
    modify column is_logic_deleted tinyint NOT NULL DEFAULT '0' COMMENT 'whether the file is logically deleted, 0-normal, 1-deleted',
    modify column is_physic_deleted tinyint NOT NULL DEFAULT '0' COMMENT 'whether the file is physically deleted, 0-normal, 1-deleted',
    drop column fs_group_id;