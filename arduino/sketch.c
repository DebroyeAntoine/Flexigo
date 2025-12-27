#include <IRremote.h>
#include <Mouse.h>

#define IR_RECEIVE_PIN 4
#define IR_SEND_PIN    9
#define BUTTON_PIN     2

#define MAX_RAW_LEN 100
#define RECV_TIMEOUT_MS 5000   // écoute IR pendant 5 s

// ===== Etat réception =====
bool recvMode = false;
unsigned long recvStartTime = 0;

// ===== KeepAlive =====
unsigned long lastSerialKA = 0;
unsigned long lastHIDKA    = 0;

void setup() {
    Serial.begin(9600);
    delay(2000);                 // IMPORTANT Win11 (remplace while(!Serial))

    pinMode(BUTTON_PIN, INPUT_PULLUP);

    IrReceiver.begin(IR_RECEIVE_PIN, ENABLE_LED_FEEDBACK);
    IrSender.begin(IR_SEND_PIN);

    Mouse.begin();               // HID souris

    Serial.println("READY");
}

void loop() {

    // ===== API SERIE =====
    if (Serial.available()) {
        String line = Serial.readStringUntil('\n');
        line.trim();

        if (line == "recvIR") {
            recvMode = true;
            recvStartTime = millis();

            IrReceiver.resume();   // purge trames précédentes
            Serial.println("OK:RECV");
        }
        else if (line.startsWith("sendIR:")) {
            parseAndSendIR(line.substring(7));
        }
        else {
            Serial.println("ERR:UNKNOWN_CMD");
        }
    }

    // ===== Réception IR continue pendant 5 s =====
    if (recvMode && IrReceiver.decode()) {

        if (IrReceiver.decodedIRData.protocol != UNKNOWN) {
            Serial.print("IR:");
            Serial.print(getProtocolString(IrReceiver.decodedIRData.protocol));
            Serial.print(",");
            Serial.print(IrReceiver.decodedIRData.address, HEX);
            Serial.print(",");
            Serial.println(IrReceiver.decodedIRData.command, HEX);
        }
        else {
            Serial.print("IR:RAW");
            uint16_t len = min(IrReceiver.decodedIRData.rawlen - 1, MAX_RAW_LEN);
            for (uint16_t i = 0; i < len; i++) {
                Serial.print(",");
                Serial.print(IrReceiver.irparams.rawbuf[i + 1] * MICROS_PER_TICK);
            }
            Serial.println();
        }

        IrReceiver.resume();   // prêt pour le prochain code
    }

    // ===== Fin écoute après 5 s =====
    if (recvMode && (millis() - recvStartTime > RECV_TIMEOUT_MS)) {
        recvMode = false;
        Serial.println("ERR:TIMEOUT");
    }

    // ===== Bouton physique → clic souris HID =====
    static bool lastBtn = HIGH;
    bool btn = digitalRead(BUTTON_PIN);

    if (lastBtn == HIGH && btn == LOW) {
        // Réveil PC : micro mouvement + clic
        Mouse.move(1, 0, 0);
        delay(10);
        Mouse.move(-1, 0, 0);
        delay(10);

        Mouse.press(MOUSE_LEFT);
        delay(20);

        Serial.println("BTN:CLICK");
        delay(300); // anti-rebond simple
    }
    lastBtn = btn;

    // ===== KEEPALIVE SERIAL =====
    if (millis() - lastSerialKA > 2000) {
        lastSerialKA = millis();
    }

    // ===== KEEPALIVE HID SOURIS =====
    if (millis() - lastHIDKA > 3000) {
        //    Mouse.move(0, 0, 0);   // rapport HID neutre
        lastHIDKA = millis();
    }
}

// ================= FONCTION TX =================

void parseAndSendIR(String payload) {

    int p1 = payload.indexOf(',');
    int p2 = payload.indexOf(',', p1 + 1);

    if (p1 < 0) {
        Serial.println("ERR:FORMAT");
        return;
    }

    String proto = payload.substring(0, p1);
    proto.toUpperCase();

    // ===== RAW =====
    if (proto == "RAW") {
        uint16_t raw[MAX_RAW_LEN];
        uint16_t rawLen = 0;

        int idx = p1 + 1;
        while (idx >= 0 && rawLen < MAX_RAW_LEN) {
            int next = payload.indexOf(',', idx);
            String val = (next < 0)
                ? payload.substring(idx)
                : payload.substring(idx, next);
            raw[rawLen++] = val.toInt();
            idx = (next < 0) ? -1 : next + 1;
        }

        Serial.println("OK:SEND");
        IrSender.sendRaw(raw, rawLen, 38);
        return;
    }

    if (p2 < 0) {
        Serial.println("ERR:FORMAT");
        return;
    }

    uint16_t addr = strtoul(payload.substring(p1 + 1, p2).c_str(), NULL, 16);
    uint16_t cmd  = strtoul(payload.substring(p2 + 1).c_str(), NULL, 16);

    Serial.println("OK:SEND");

    if      (proto == "NEC")      IrSender.sendNEC(addr, cmd, 0);
    else if (proto == "SAMSUNG")  IrSender.sendSamsung(addr, cmd, 0);
    else if (proto == "SONY")     IrSender.sendSony(cmd, 12, 2);
    else Serial.println("ERR:UNKNOWN_PROTO");
}
