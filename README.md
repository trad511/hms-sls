# System Layout Service (SLS)

The System Layout Service (SLS) holds information about a Cray Shasta system "as configured".  That is, it contains information about how the system was designed.  SLS reflects the system if all hardware were present and working.  SLS does not keep track of hardware state or identifiers; for that type of information, see Hardware State Manager (HSM).

## How it Works

SLS is mostly an API to allow access to a database.  Users make queries on the basis of xname and are given back information about the system.  Each xname is matched with corresponding parent and child information, allowing the simple traversal of the entirety of the system.

## Configuration

TBD.  Initial configuration will be uploaded by an init container.

## SLS CT Testing

This repository builds and publishes hms-sls-ct-test RPMs along with the service itself containing tests that verify SLS on the
NCNs of live Shasta systems. The tests require the hms-ct-test-base RPM to also be installed on the NCNs in order to execute.
The version of the test RPM installed on the NCNs should always match the version of SLS deployed on the system.

## More information

* [SLS design documentation](https://connect.us.cray.com/confluence/display/CASMHMS/System+Layout+Service+%28SLS%29+Design+Documentation)

## Future Work

TBD.