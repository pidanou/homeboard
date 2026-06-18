export type Member = {
	user_id: string;
	name: string;
	email: string;
	role: string;
	virtual?: boolean;
};

export type VirtualMember = {
	id: string;
	family_id: string;
	name: string;
};

export type AppCategory = {
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
	important: boolean;
	assigned_to?: string;
	end_date?: string;
	start_date?: string;
	category_id?: string;
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
	category_id?: string;
};

export type Filter = 'all' | 'tasks' | 'events' | 'done';

export type AppList = {
	id: string;
	family_id: string;
	name: string;
};

export type AppListItem = {
	id: string;
	list_id: string;
	name: string;
	checked: boolean;
	created_at: string;
	checked_at?: string;
};
