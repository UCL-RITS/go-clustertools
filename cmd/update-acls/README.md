# update-acls

This is a tool intended to synchronise various forms of user access-control list.

As-is, it can read from:

 - Active Directory groups
 - UNIX groups
 - text lists
 - Gold projects (via `glsproject`)
 - UCL's particular way of including departments in Active Directory, which may not be universal
 - Sun Grid Engine ACLs (via `qconf`)

And change:

 - text lists
 - Gold projects (via `gchproject`)
 - Sun Grid Engine ACLs (via `qconf`)


## CLI

```
$ ./update-acls -h
Usage of ./update-acls:
  -config string
    	path to config file (default "./update-acls.conf")
  -no-targets
    	skip the output phase, only expand lists
  -show-lists
    	print the built lists after expansion
```

## Configuration

The configuration is specified in YAML and looks something like this:

```
ad_options:
    server_url: "ldaps://ad.ucl.ac.uk"
    bind_username: "AD\\sa-roleuser01"
    bind_password: "hunter2"
    base_dn: "DC=ad,DC=ucl,DC=ac,DC=uk"
    allow_insecure: yes
lists:
    - name: "Special Users"
      description: "Users allowed to use the special part of the cluster"
      include:
        ad_groups: ["economics-all","rescfserv-all"]
        users: ["ccspapp"]
        unix_groups: ["ccsprcop"]
      exclude:
        users: ["cceabba"]
      filter:
        sge_acls: ["Open"]
      destinations:
        text_list_files: ["./special_example.txt"]
```

This file is also in `etc/example.conf`.

Text files are expected to be new-line-separated.

Order of the sections is not important but the sections cannot occur more than once in a single list entry.

No error will occur if your AD/LDAP options are incorrect or missing if you don't specify any AD groups to include. (i.e. the only checks are made when queries are made.)

If any errors are detected during list expansion, no changes will be made to any destination.

All lists are expanded before any changes are made.

## List Expansion

The list of users is made from expanding all the `include` entries, removing all the `exclude` entries, then removing all entries that aren't in `filter`. 
So, this example:

 - expands the AD groups `economics-all` and `rescfserv-all` into a list of users
 - expands the UNIX group `ccsprcop` into a list of users and adds that
 - adds the literally specified user `ccspapp`
 - removes the literally specified user `cceabba`
 - removes any user that isn't in the `Open` SGE ACL
 - writes the list of users to a text file, `./special_example.txt`

