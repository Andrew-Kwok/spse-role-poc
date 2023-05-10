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


Available roles: `{"Admin PPE", "Admin Agency", "Verifikator", "Helpdesk", "PPK", "KUPBJ", "Anggota Pokmil", "PP", "Auditor"}` 
