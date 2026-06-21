import type { CapacitorConfig } from '@capacitor/cli';

const config: CapacitorConfig = {
	appId: 'com.noudapi.familyboard',
	appName: 'Family Board',
	webDir: 'build',
	ios: {
		minVersion: '15.0',
	},
};

export default config;
