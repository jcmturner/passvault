# passvault

A simple password management tool that encrypts passwords with a master
passphrase and stores them locally on disk.

The password is copied to the clipboard by default.

## Encryption
The encryption used is AES256-GCM.

Key derivation uses pbkdf2 with 1000 iterations.

## Usage
```bash
./passvault
  -add
        Add entry
  -del
        Delete a secret entry
  -json
        Output in JSON format
  -list
        List the names of all the secret entries
  -name string
        Name of the secret entry
  -print
        Print details to stdout
  -session
        Run as non-exiting read only session
  -vault string
        Specify path to the vault directory (default "/Users/turnerj/passvault")
```

### Adding Entries
```bash
./passvault -add
```
You will first be prompted for a passphrase to generate the encryption key from.
This can be the same for all entries.
You will be prompted for a name for the entry. This name must be unique or
the existing entry under that name will be overwritten.

```
Enter passphrase: 
Enter name of secret entry: 
Username: 
Password: 
Secret stored successfully
```

### Delete Entries
```bash
./passvault -del
```
You will be prompted for the name of the secret to delete.

To confirm the deletion you must enter the passphrase that was used to encrypt it.
This must be the correct passphrase to decrypt the entry.
```
Enter name of secret entry: 
Confirm deletion of secret entry test with passphrase
Enter passphrase: 
Entry test deleted successfully
```

### Access Entries
This is the default behaviour.
```
./passvault
```
You will b prompted for the passphrase and then the name of the entry.
By defaut the password value is copied to the clipboard.
```
Enter passphrase: 
Enter name of secret entry: 
Name: 
Username: 
Password: ****
Password copied to clipboard.
```
If you just want the details to be printed to the screen pass the ``-print`` switch.
To have this printed in JSON form also pass the ``-json`` switch.

### List Entries
```
./passvault -list
```
A list of all the entry names is returned.

### Session Mode
The session mode enables you to enter the passphase once and then repeatedly asked for entries
and have their password copied to the clipboard.

This assumes that all the entries have used the same passphrase.

The passphrase is re-requested every 8 hours.

```
./passvault -session
Enter passphrase: 
------
Enter name of secret entry: test
Name: test
Username: jt
Password: ****
Password copied to clipboard.
------
Enter name of secret entry: test
Name: test
Username: jt
Password: ****
Password copied to clipboard.
------
```

## Vault Location
By default the vault is stored on disk in a directory call ``passvault`` in the user's
home directory.
This can be overridden with the ``-vault`` argument.

The vault uses ``github.com/prologic/bitcask`` to manage the storage.

No sensitive values are saved to disk unencrypted.

## At Your Own Risk
Use this tool at your own risk.
If you forget your passphrase(s) there is no way to retrieve the values.
The author accepts no liability for data lost as a result of using this tool.

## Build
```bash
go build -o passvault github.com/jcmturner/passvault/cmd
```