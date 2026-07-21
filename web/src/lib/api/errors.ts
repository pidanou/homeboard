import * as m from '$lib/paraglide/messages.js';

/** Maps stable backend error codes (see docs/specs/i18n.md) to localized message functions. */
const errorMessages: Record<string, () => string> = {
	accept_invite_failed: m.error_accept_invite_failed,
	add_item_failed: m.error_add_item_failed,
	admin_required: m.error_admin_required,
	cancel_occurrence_failed: m.error_cancel_occurrence_failed,
	cannot_change_own_role: m.error_cannot_change_own_role,
	cannot_demote_last_admin: m.error_cannot_demote_last_admin,
	cannot_remove_self: m.error_cannot_remove_self,
	check_households_failed: m.error_check_households_failed,
	clear_checked_items_failed: m.error_clear_checked_items_failed,
	complete_login_failed: m.error_complete_login_failed,
	create_category_failed: m.error_create_category_failed,
	create_event_failed: m.error_create_event_failed,
	create_family_failed: m.error_create_family_failed,
	create_invite_failed: m.error_create_invite_failed,
	create_list_failed: m.error_create_list_failed,
	create_subscription_failed: m.error_create_subscription_failed,
	create_task_failed: m.error_create_task_failed,
	delete_category_failed: m.error_delete_category_failed,
	delete_event_failed: m.error_delete_event_failed,
	delete_invite_failed: m.error_delete_invite_failed,
	delete_item_failed: m.error_delete_item_failed,
	delete_list_failed: m.error_delete_list_failed,
	delete_subscription_failed: m.error_delete_subscription_failed,
	delete_task_failed: m.error_delete_task_failed,
	export_not_found: m.error_export_not_found,
	family_not_found: m.error_family_not_found,
	file_too_large: m.error_file_too_large,
	forbidden: m.error_forbidden,
	generate_export_token_failed: m.error_generate_export_token_failed,
	get_members_failed: m.error_get_members_failed,
	get_virtual_members_failed: m.error_get_virtual_members_failed,
	ids_required: m.error_ids_required,
	image_required: m.error_image_required,
	import_calendar_failed: m.error_import_calendar_failed,
	invalid_credentials: m.error_invalid_credentials,
	invalid_current_password: m.error_invalid_current_password,
	invalid_date_range: m.error_invalid_date_range,
	invalid_end_at: m.error_invalid_end_at,
	invalid_end_date: m.error_invalid_end_date,
	invalid_or_expired_code: m.error_invalid_or_expired_code,
	invalid_push_subscription: m.error_invalid_push_subscription,
	invalid_request: m.error_invalid_request,
	invalid_reset_link: m.error_invalid_reset_link,
	invalid_role: m.error_invalid_role,
	invalid_start_at: m.error_invalid_start_at,
	invalid_start_date: m.error_invalid_start_date,
	invite_already_used: m.error_invite_already_used,
	invite_expired: m.error_invite_expired,
	invite_no_email: m.error_invite_no_email,
	invite_not_found: m.error_invite_not_found,
	issue_token_failed: m.error_issue_token_failed,
	list_categories_failed: m.error_list_categories_failed,
	list_events_failed: m.error_list_events_failed,
	list_families_failed: m.error_list_families_failed,
	list_invites_failed: m.error_list_invites_failed,
	list_items_failed: m.error_list_items_failed,
	list_lists_failed: m.error_list_lists_failed,
	list_subscriptions_failed: m.error_list_subscriptions_failed,
	list_tasks_failed: m.error_list_tasks_failed,
	missing_file: m.error_missing_file,
	name_and_url_required: m.error_name_and_url_required,
	name_required: m.error_name_required,
	no_password_set: m.error_no_password_set,
	not_found: m.error_not_found,
	oidc_exchange_failed: m.error_oidc_exchange_failed,
	oidc_flow_expired: m.error_oidc_flow_expired,
	oidc_invalid_claims: m.error_oidc_invalid_claims,
	oidc_invalid_nonce: m.error_oidc_invalid_nonce,
	oidc_invalid_state: m.error_oidc_invalid_state,
	oidc_missing_code: m.error_oidc_missing_code,
	oidc_start_failed: m.error_oidc_start_failed,
	password_login_disabled: m.error_password_login_disabled,
	password_too_short: m.error_password_too_short,
	process_request_failed: m.error_process_request_failed,
	read_file_failed: m.error_read_file_failed,
	registration_closed: m.error_registration_closed,
	registration_failed: m.error_registration_failed,
	remove_subscription_failed: m.error_remove_subscription_failed,
	rename_list_failed: m.error_rename_list_failed,
	reorder_items_failed: m.error_reorder_items_failed,
	reorder_lists_failed: m.error_reorder_lists_failed,
	reorder_tasks_failed: m.error_reorder_tasks_failed,
	reset_password_failed: m.error_reset_password_failed,
	revoke_export_token_failed: m.error_revoke_export_token_failed,
	save_subscription_failed: m.error_save_subscription_failed,
	single_household_mode: m.error_single_household_mode,
	storage_error: m.error_storage_error,
	streaming_unsupported: m.error_streaming_unsupported,
	subscription_not_found: m.error_subscription_not_found,
	sync_failed: m.error_sync_failed,
	target_not_member: m.error_target_not_member,
	task_not_found: m.error_task_not_found,
	title_required: m.error_title_required,
	unauthorized: m.error_unauthorized,
	unsupported_image_type: m.error_unsupported_image_type,
	unsupported_locale: m.error_unsupported_locale,
	unsupported_reminder_offset: m.error_unsupported_reminder_offset,
	unsupported_time_format: m.error_unsupported_time_format,
	update_category_failed: m.error_update_category_failed,
	update_event_failed: m.error_update_event_failed,
	update_failed: m.error_update_failed,
	update_item_failed: m.error_update_item_failed,
	update_occurrence_failed: m.error_update_occurrence_failed,
	update_task_failed: m.error_update_task_failed
};

/**
 * Resolves an HTTP error response body into a localized message. Backend
 * handlers return `{"error": "<code>"}`; unmigrated handlers (or non-JSON
 * bodies) fall back to the raw response text, then the status text.
 */
export async function resolveErrorMessage(res: Response): Promise<string> {
	const text = await res.text();
	if (text) {
		try {
			const body = JSON.parse(text);
			if (typeof body?.error === 'string' && errorMessages[body.error]) {
				return errorMessages[body.error]();
			}
		} catch {
			// not JSON — an unmigrated handler's raw text response
		}
	}
	return text || res.statusText;
}
