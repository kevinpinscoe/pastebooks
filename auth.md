# Authentication

In Version 1.0 authentication is using email and a password.

Actually there is no email verification backend so email can be any text really just a name or string of text. 

I am considering stronger authentication methods for 2.0 but for local private networks this is working.

Note that currently the app has no change password functionality.

To re-register a user (lost password):

Note you will lose all your charms and books.

`docker exec -it pastebooks-db mysql -uroot -prootpass charmsdb`

Confirm it exists:

`mysql> SELECT id, email, created_at FROM users WHERE email\='some.user@someemail.com'; 

Remove it:

`mysql> DELETE FROM users WHERE email\='some.user@someemail.com';`

Now visit http://host/pastebooks and re-register.

To change the password:

Set a new password hash without deleting books/charms

This preserves the id, so existing books.owner_id still matches.

Generate a bcrypt hash for your new passcode using a one-off container: 

```
NEWPASS\='put-your-new-passcode-here'
HASH\=$(docker run -rm golang:1.24-bookworm bash -lc \
'cat > /tmp/bcrypt.go <<'"'"'EOF'"'"'
package main
import (
"fmt"
"os"
"golang.org/x/crypto/bcrypt"
)
func main() {
h, _ := bcrypt.GenerateFromPassword([]byte(os.Args[1]), bcrypt.DefaultCost)
fmt.Print(string(h))
}
EOF
cd /tmp
go mod init tmp >/dev/null 2>&1
go get golang.org/x/crypto/bcrypt >/dev/null 2>&1
go run /tmp/bcrypt.go "'"$NEWPASS"'"
')
echo "$HASH" 
```

Update the user's pass_hash (store as bytes; MySQL will accept the bcrypt string): 

```
docker exec -i pastebooks-db mysql -uroot -prootpass charmsdb <<SQL
UPDATE users
SET pass_hash = '${HASH}'
WHERE email = 'some.user@someemail.com';
```


