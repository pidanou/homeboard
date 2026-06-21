import type { CapacitorConfig } from '@capacitor/cli';

const config: CapacitorConfig = {
	appId: 'com.noudapi.homeboard',
	appName: 'Homeboard',
	webDir: 'build',
	ios: {
		minVersion: '15.0',
	},
};

export default config;
