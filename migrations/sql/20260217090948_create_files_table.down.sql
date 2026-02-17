DROP INDEX IF EXISTS idx_file_favorites_file_id;
DROP INDEX IF EXISTS idx_file_favorites_user_id;
DROP INDEX IF EXISTS idx_file_last_viewed_user_viewed;
DROP INDEX IF EXISTS idx_file_views_user_id;
DROP INDEX IF EXISTS idx_file_views_file_id;
DROP INDEX IF EXISTS idx_file_grants_grantee;
DROP INDEX IF EXISTS idx_file_grants_file_id;
DROP INDEX IF EXISTS idx_files_user_id;

DROP TABLE IF EXISTS file_favorites;
DROP TABLE IF EXISTS file_last_viewed;
DROP TABLE IF EXISTS file_views;
DROP TABLE IF EXISTS file_grants;
DROP TABLE IF EXISTS files;

DROP TYPE IF EXISTS permission;
DROP TYPE IF EXISTS mime_type;
DROP TYPE IF EXISTS upload_status;
DROP TYPE IF EXISTS grantee_type;