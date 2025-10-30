# Database management

The database by default runs in a container, however it can be configured to use a persisted MySQL database instance.

If using the container you will not need to make exports and imports of the data on each restart. However as a course of backups you should back the data up periodically. 

## Backup and restore

### Backup

Dumps only the charmsdb schema+data to a file on your host

```
docker exec -i pastebooks-db \
  mysqldump -uroot -prootpass --routines --triggers --single-transaction \
  charmsdb > backup-charmsdb-$(date +%F).sql
```

### Restore

```
# Restore into charmsdb (must exist). You can re-use your schema.sql or a full dump:
docker exec -i pastebooks-db \
  mysql -uroot -prootpass charmsdb < backup-charmsdb-2025-10-23.sql
```

## Managing users

Users have an email field and a passcode field.

### Listing users

```
docker exec -i pastebooks-db \
  mysql -uroot -prootpass -N -e "USE charmsdb; SELECT id,email,created_at FROM users;"
```

### Deleting users

IMPORTANT! deleting a user deletes books and charms created by that user due to 
ON DELETE CASCADE.

Delete an app user by email:

```
EMAIL="kevin.inscoe@gmail.com"
docker exec -i pastebooks-db \
  mysql -uroot -prootpass -e "USE charmsdb; DELETE FROM users WHERE email='${EMAIL}';"
```

# Create a test user

Generate a bcrypt hash with your app or `htpasswd -bnBC 10 '' pass | cut -d: -f2`.

```
docker exec -i pastebooks-db mysql -uroot -prootpass -D charmsdb -e "
  INSERT INTO users (id,email,pass_hash)
  VALUES (UUID(), 'you@example.com', '\$2y\$10\$replace_with_bcrypt_hash');
"
```

## Sanity checks

Confirm DB exists and tables are present:

```
docker exec -i pastebooks-db \
  mysql -uroot -prootpass -e "SHOW DATABASES; USE charmsdb; SHOW TABLES;"
```

## Database passwords

Becayse we are running mytsql as a root user a password does not need to be passed into MySQL. However if you create a least privilege user (which is not recommended and s not generally needed)

Create a least-privilege app user (example):

```
docker exec -i pastebooks-db \
  mysql -uroot -prootpass -e "
    CREATE USER IF NOT EXISTS 'pastebooks'@'%' IDENTIFIED BY 'strongpass';
    GRANT SELECT,INSERT,UPDATE,DELETE ON charmsdb.* TO 'pastebooks'@'%';
    FLUSH PRIVILEGES;"
```

Drop a MySQL account:

```
docker exec -i pastebooks-db \
  mysql -uroot -prootpass -e "DROP USER IF EXISTS 'pastebooks'@'%'; FLUSH PRIVILEGES;"

```

Avoid putting passwords in your shell history: use environment variables or a .my.cnf inside the container (or docker exec -e MYSQL_PWD=...).

## Best practices

Keep mysqldump --single-transaction for consistent hot backups on InnoDB.

Currently using collation utf8mb4_0900_ai_ci which implies MySQL 8 so stick to mysql:8.0 images unless it gets upgraded here. When that happens upgrade instructions will appear in release notes yet to be created.

