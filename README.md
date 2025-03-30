# Golang expense recording web server

This project can be used to store expenses in lightweight sqlite db with simple web interface and can be run in a docker container. 
You can see list of project's future improvements in [Future improvements](#future-improvements)

---

## Full Stack Overview

### What Has Been Used

- **[fiber](github.com/gofiber/fiber/v2)** – Web framework for REST endpoints
- **[Docker](https://www.docker.com/)** – Containerization
- **[sqlite](github.com/mattn/go-sqlite3)** – Go driver for sqlite
- **html/template** – Render html UI
---

## How to Run

### Option A: Docker

1. Clone the repository.
2. Navigate to the project's root directory.
3. Start with:

   ```bash
   docker build -t expenses-img .
   ```

   This creates docker image named expenses-img.

   ```bash
   docker run -d -p 8000:5000 --name expenses-cont expenses-img
   ```

   This runs docker container named expenses-cont using internal docker port 5000 and host port 8000.

**Verify:**

- Open in browser at [http://localhost:8000/expense](http://localhost:8000/expense).

### Option B: Run on the host machine

- Compile source code: 

  ```bash
  go build -o expenses main.go
  ```

**Verify:**

- Open in browser at [http://localhost:8000/expense](http://localhost:8000/expense).

## UI

- /expense – See all expenses in specified dates (current month by default)
- /expense – Update or remove any expense
- /expense/stat – See statistics about expenses in specified dates (current month by default)
- /expense/add – Add expense one at a time of use a batch format. You should use "Проверка чеков ФНС России" app and its json output. Use a comma sign to separate categories
- /expense/search – Search expense by name

## Future improvements

- Change batched expenses to new yaml schema
- Add **[htmx](https://htmx.org/)** – Dependencies free js lib giving access to AJAX and more
- Add DB backup in UI
- Add filter by category in /expense
- Add expenses representation of choosen category in /expense/stat
- Add a category comparison per month in /expense/stat
- Add found expenses comparison by prices and dates
- Improve UI with css