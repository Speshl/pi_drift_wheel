
#define GYRO_PIN 2
#define FEEDBACK_PIN A7

#define FEEDBACK_SMOOTHING 5

#define MIN_FEEDBACK 50 //smoothed
#define MID_FEEDBACK 500 //smoothed calc
#define MAX_FEEDBACK 950 //smoothed

#define MIN_PWM 500
#define MID_PWM 1500
#define MAX_PWM 2500


volatile long GyroLastChangeTime = 0;
volatile long GyroChangeTime = 0;
volatile long GyroPulseTime = 0;

int usedPulseTime = MID_PWM;

int16_t loopValue = -180;

int FeedbackToAvg[FEEDBACK_SMOOTHING];

void setup() {
  //set up the serial monitor, pin mode, and external interrupt.
  Serial.begin(460800);
  // while(!Serial){
  //   ;
  // }
  Serial1.begin(420000); 

  delay(100);
  Serial.println("ready");
}

void loop() {
  int totalFeedback = 0;
  int feedbackToShift = analogRead(FEEDBACK_PIN);
  for(int i=0; i<FEEDBACK_SMOOTHING; i++){
    totalFeedback += feedbackToShift;
    int temp = FeedbackToAvg[i];
    FeedbackToAvg[i] = feedbackToShift;
    feedbackToShift = temp;
  }
  int smoothedFeedback = totalFeedback / FEEDBACK_SMOOTHING;
  Serial.print("Smoothed: ");
  Serial.println(smoothedFeedback);

  byte pitchMSB = (smoothedFeedback >> 8) & 0xFF; // Most significant byte (MSB)
  byte pitchLSB = smoothedFeedback & 0xFF; // Least significant byte (LSB)

  uint8_t framePreCalc[] = {0x1E,pitchMSB,pitchLSB,0x00,0x00,0x00,0x00};
  uint8_t crc8 = GenerateCrc8Value(framePreCalc, sizeof(framePreCalc));
  uint8_t framePostCalc[] = {0xC8,0x08,0x1E,pitchMSB,pitchLSB,0x00,0x00,0x00,0x00,crc8};
  Serial1.write(framePostCalc, sizeof(framePostCalc));
}

int MapToRange(int value, int min, int max, int minReturn, int maxReturn) {
	int mappedValue = (maxReturn-minReturn)*(value-min)/(max-min) + minReturn;

	if(mappedValue > maxReturn){
		return maxReturn;
	} else if(mappedValue < minReturn){
		return minReturn;
	} else {
		return mappedValue;
	}
}

uint8_t crc8_dvb_s2(uint8_t crc, uint8_t a) {
  crc = crc ^ a;
  for (uint8_t ii = 0; ii < 8; ++ii) {
    if (crc & 0x80) {
      crc = (crc << 1) ^ 0xD5;
    } else {
      crc = crc << 1;
    }
  }
  return crc & 0xFF;
}

uint8_t GenerateCrc8Value(uint8_t* data, size_t dataSize) {
  uint8_t crc = 0;
  for (size_t i = 0; i < dataSize; ++i) {
    crc = crc8_dvb_s2(crc, data[i]);
  }
  return crc;
}

bool ValidateFrame(uint8_t* frame, size_t frameSize) {
  uint8_t crc = GenerateCrc8Value(frame + 2, frameSize - 3);
  return crc == frame[frameSize - 1];
}

