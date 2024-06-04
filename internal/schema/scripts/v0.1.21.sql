alter table bookmark
    drop key `uk_md5`,
    drop key `idx_user_no`,
    add key `uk_user_no_md5` (user_no, md5);

alter table bookmark_blacklist
    add key `idx_name` (name);