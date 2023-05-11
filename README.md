# spse-role-poc

Fill `MGMT_ACCESS_TOKEN=` in `.env`. This can be obtained from Auth0 > APIs > Auth0 Management API > API Explorer.

Start the API by calling `go run .`. This will starts the API
To create a user, send a `GET` request to `localhost:3000/create` with request body
```
{
    "email": "{user_email}",
    "password": "{user_password}",
    "roles": ["{user_role_1_name}, {user_role_2_name}, ..."]
}
```


Available roles: `{"A1:Admin PPE", "A1:Admin Agency", "A1:Verifikator", "A1:Helpdesk", "A1:PPK", "A1:KUPBJ", "A1:Anggota Pokmil", "A1:PP", "A1:Auditor", "A2:Admin PPE", ..., "A3:Auditor"}` 
