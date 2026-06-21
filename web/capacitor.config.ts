import type { CapacitorConfig } from '@capacitor/cli';

const liveReloadUrl = process.env.LIVE_RELOAD_URL;

const config: CapacitorConfig = {
	appId: 'com.noudapi.homeboard',
	appName: 'Homeboard',
	webDir: 'build',
	ios: {
		minVersion: '15.0',
	},
	...(liveReloadUrl ? { server: { url: liveReloadUrl, cleartext: true } } : {}),
};

export default config;
