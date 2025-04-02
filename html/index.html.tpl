<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset="UTF-8">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css" rel="stylesheet"
        integrity="sha384-QWTKZyjpPEjISv5WaRU9OFeRpok6YctnYmDr5pNlyT2bRjXh0JMhjY6hW+ALEwIH" crossorigin="anonymous">
    <script src="https://unpkg.com/htmx.org@2.0.4"></script>
    <title>SQLite MT</title>
</head>

<body>
    <div class="row">
        <div class="col">
            <h2>Create User</h2>
            <div>
                <a href="#" hx-get="/newuser" hx-target="#createUser">CreateUser</a>
            </div>
        </div>
        <!-- <div class="col">
            
        </div> -->
        <div class="col">
            <h2>Add New Task</h2>
            <div id="addTaskForm">
                {{template "addTaskForm"}}
            </div>

        </div>
    </div>

</body>

</html>