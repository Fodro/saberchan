export type Captcha = {
	token: string;
	input: string;
};

export type CaptchaRes = {
	passed: boolean;
}