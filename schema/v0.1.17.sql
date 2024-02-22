USE vfm;

-- warn: this will rebuilt the table, may take a while
ALTER TABLE file_info
    DROP INDEX name,
    ADD FULLTEXT name_idx (name);