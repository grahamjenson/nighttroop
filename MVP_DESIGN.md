# NightTroop Design

Night**Troop** hosts many (10kish) Night**Scout** servers. 
The main goal is to reduce the complexity of setting up NightScout to a single **FREE** click.

# On Boarding

1. Go to nt01au.t1d.nz (something like nt<instance #><zone>.t1d.nz)
2. Click the "Create Site" button
3. Copy the **Site Data** and store it somewhere safe

# Getting data into Nighttroop

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
