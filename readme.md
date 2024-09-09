## Numeris Task API
This Backend Server is built with Go, Chi as the web framework of choice, and Postgres is Used on the Backend.

So what was not done?
1. Tests. Currently, there are no test, I am still studying and learning Go and racing the Time (which was sufficient).
2. Job Processing (Reminders would ideally be implemented. That can be implemented using dkron server especially since it is implemented in Go)

What was Done?
1. Create, Read, Update Endpoints for Invoices
2. Dashboard Endpoint for Invoices providing a summary of invoices.
3. There was  a bit of extra information such as User/Org Account Information. This was Faked or Dummified

What would I do with more time and building the software?
 Offering Holding Virtual Accounts that could/should reconcile to the business main account, As such we could hook some actions, such that when the account receives payment, the invoice gets updated eliminating the manual payment update.

Here is an example of documentation via Postman:

```https://documenter.getpostman.com/view/17138168/2sAXjSyoKM```
My choice of documentation tool is Swagger but racing the clock I could not look out for that.

I used .env to show that I understand and have experience building server_side logic. I provided an example.env file.

