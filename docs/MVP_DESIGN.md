# NightTroop Design

This document is the stories I want the user and the admin to experience when using Nighttroop for the MVP.

## Site User

### On Boarding

1. Go to nt01au.t1d.nz (something like nt<instance #><zone>.t1d.nz)
2. Click the "Create Site" button
3. Copy the **Site Data** and store it somewhere safe

### Getting data into Nighttroop

Getting the BGL and treatment data into Nightscout should be pretty easy

#### Person With Diabetes (PWD) Phone

You can upload data directly from AAPS/Trio:

1. Copy the NS URL and Token from the **Site Data**
2. Copy to PWD phone AAPS/Trio with information 

#### Dexcom/Nightscout/Libre Sync

You can also setup a sync to pull data from other service:

1. Login with Admin password and TOTP token from the **Site Data**
2. In the **Settings Menu** click the Sync settings

**Dexcom**:
1. Fill in Dexcom Username and Password
2. Save

**Nightscout**:
1. Fill in the Nightscout URL and Token of the other NS site
2. Save

**Libre**
1. Fill in the Libre information
2. Save

# Sharing Data in Nighttroop

In the **Settings Menu** go into "Devices and Users"

We have added a list of pre-configured device and user types that you might want. 


# Controlling AAPS(Trio?) with Nighttroop

AAPS can sync data like Carbs, Temp targets, Profile Changes with % written to Nighttroop

Via WebSite:

1. You log in as a user that can add treatments, like admin
1. Go to treatments page
1. Add the needed treatment

Via T1DCarer

1. Create a user for T1DCarer that can add treatments
1. Add that user to T1DCarer


# Control Level Features

1. Total Users

## User management

1. Seeing users database size
1. looking at user syncs, errors reports
1. Live reporting API limits and total violations (I want to be able to understand limits)

1. Delete/Archive a user
1. Restore a user from archive

## Job Management

1. Seeing Live Job queue
1. Seeing Job reports
1. Seeing Job errors

## Errors

1. Logging errors to Sentry
1. Some kind of live alert for if the site goes down
1. See errors as events locally
1. All 500s are things genuinly should never happen

## Events

1. I need to see events
1. Log events