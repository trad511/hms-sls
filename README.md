# System Layout Service (SLS)

The System Layout Service (SLS) holds information about a Cray Shasta system "as configured".  That is, it contains information about how the system was designed.  SLS reflects the system if all hardware were present and working.  SLS does not keep track of hardware state or identifiers; for that type of information, see Hardware State Manager (HSM).

## How it Works

SLS is mostly an API to allow access to a database.  Users make queries on the basis of xname and are given back information about the system.  Each xname is matched with corresponding parent and child information, allowing the simple traversal of the entirety of the system.

## Configuration

TBD.  Initial configuration will be uploaded by an init container.

## More information

* [SLS design documentation](https://connect.us.cray.com/confluence/display/CASMHMS/System+Layout+Service+%28SLS%29+Design+Documentation)

## Future Work

TBD.