# VaspList

This is pretty janky, honestly -- it scrapes the VASP license portal and tries to work out what's going on from various patterns.


## Credentials

To make this work, set as environment variables:

```
VASPTOOL_USERNAME
VASPTOOL_PASSWORD
VASPTOOL_LICNUM
```

The license number here is a 5-digit number rather than the form with a dash in -- I grabbed it from the internal values on the page.


## Patterns and Workings

First it has to authenticate, get the page back, and then POST back a user to add if appropriate.

There are a couple of alert forms it looks for when it tries to add people:

```
<div class="alert alert-success" role="alert">
User 'person@place.ac.uk' added to license 'License AB01-1234 5-678'

<div class="alert alert-danger" role="alert">
No user with email 'person@place.ac.uk' found!

<div class="alert alert-danger" role="alert">
User 'person@place.ac.uk' already member of license 'License AB01-1234 5-678'
```

The workings of this have been created based on response analysis of the website, so it could be incorrect or break at any time if they change it.


