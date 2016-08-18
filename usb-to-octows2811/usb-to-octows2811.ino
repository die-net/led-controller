
#define USE_OCTOWS2811
#include<OctoWS2811.h>
#include<FastLED.h>

#define NUM_LEDS_STRIP_A 514
#define NUM_LEDS_STRIP_B 370
#define NUM_LEDS_STRIP_C 238
#define NUM_LEDS_STRIP_D 102
#define NUM_PIXELS_PER_SUPPLY (NUM_LEDS_STRIP_A + NUM_LEDS_STRIP_B + NUM_LEDS_STRIP_C + NUM_LEDS_STRIP_D)
#define NUM_SUPPLIES 2
#define NUM_LEDS (NUM_PIXELS_PER_SUPPLY * NUM_SUPPLIES)

CRGB leds[NUM_LEDS];

#define MAX_BRIGHTNESS  255

#define MAX_SUPPLY_MW 240000  // 300W * 80% * 1000MW/W
#define RED_MW_PER_LED 119
#define GREEN_MW_PER_LED 92
#define BLUE_MW_PER_LED 89

// Pin layouts on the teensy 3:
// OctoWS2811: 2,14,7,8,6,20,21,5

// Pin 13 has the LED on Teensy 3.0
// give it a name:
#define STATUS_LED 13
int frame_count = 0;

void setup() {
  pinMode(STATUS_LED, OUTPUT);
  digitalWrite(STATUS_LED, HIGH);

  Serial.begin(115200);
  FastLED.addLeds<OCTOWS2811>(leds, NUM_LEDS_STRIP_A);
  FastLED.addLeds<OCTOWS2811>(leds, NUM_LEDS_STRIP_B);
  FastLED.addLeds<OCTOWS2811>(leds, NUM_LEDS_STRIP_C);
  FastLED.addLeds<OCTOWS2811>(leds, NUM_LEDS_STRIP_D);
  FastLED.addLeds<OCTOWS2811>(leds, NUM_LEDS_STRIP_A);
  FastLED.addLeds<OCTOWS2811>(leds, NUM_LEDS_STRIP_B);
  FastLED.addLeds<OCTOWS2811>(leds, NUM_LEDS_STRIP_C);
  FastLED.addLeds<OCTOWS2811>(leds, NUM_LEDS_STRIP_D);

  FastLED.setBrightness(MAX_BRIGHTNESS);
  leds[0] = CRGB::Red;
  leds[1] = CRGB::Green;
  leds[2] = CRGB::Blue;
  FastLED.show();

  delay(500);
  digitalWrite(STATUS_LED, LOW);

  // Now turn the LED off, then pause
  leds[0] = CRGB::Black;
  leds[1] = CRGB::Black;
  leds[2] = CRGB::Black;
  FastLED.show();
  delay(500);

  digitalWrite(STATUS_LED, HIGH);
}

long leds_mw_per_supply() {
  long max_mw = 0;
  for (int led = 0; led < NUM_LEDS; ) {
    long red = 0;
    long green = 0;
    long blue = 0;
    for (int e = led + NUM_PIXELS_PER_SUPPLY; led < e; led++) {
      red += leds[led].red;
      green += leds[led].green;
      blue += leds[led].blue;
    }
    long mw = (red * RED_MW_PER_LED + green * GREEN_MW_PER_LED + blue * BLUE_MW_PER_LED) / 255;
    max_mw = max(max_mw, mw);
  }
  return max_mw;
}

void receive_frame() {
  int startChar = Serial.read();
  if (startChar != '*') {
    return;
  }

  int brightness = Serial.read();
  if (brightness < 0) {
    return;
  }
  // Make sure we don't exceed hardcoded limit.
  brightness = min(brightness, MAX_BRIGHTNESS);

  // read three bytes: r, g, and b.
  Serial.readBytes( (char*)leds, NUM_LEDS * 3);

  // Limit brightness based on guess of power draw per supply.
  long mw = leds_mw_per_supply();
  if (mw > MAX_SUPPLY_MW) {
    brightness = brightness * MAX_SUPPLY_MW / mw;
    frame_count += 7;  // Make status LED blink faster.
  }

  FastLED.show(brightness);

  frame_count++;
  digitalWrite(STATUS_LED, (frame_count & 0x80) ? HIGH : LOW);

  Serial.print(brightness);
  Serial.print("\t");
  Serial.println(mw);
}

void loop() {
  if (Serial.available() > 0) {
    receive_frame();
  }
}

