export type Member = {
	user_id: string;
	name: string;
	email: string;
	role: string;
};

export type AppLabel = {
	id: string;
	family_id: string;
	name: string;
	color: string;
};

export type Task = {
	id: string;
	title: string;
	description?: string;
	status: string;
	priority: string;
	assigned_to?: string;
	end_date?: string;
	start_date?: string;
	label_ids?: string[];
};

export type CalEvent = {
	id: string;
	title: string;
	description?: string;
	location?: string;
	start_at: string;
	end_at: string;
	all_day: boolean;
	attendee_ids?: string[];
	label_ids?: string[];
};

export type Filter = 'all' | 'tasks' | 'events' | 'done';
