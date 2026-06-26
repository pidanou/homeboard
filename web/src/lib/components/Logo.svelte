<script lang="ts">
	let { size = 28, class: cls = '' }: { size?: number; class?: string } = $props();
</script>

<svg
	xmlns="http://www.w3.org/2000/svg"
	width={size}
	height={size}
	viewBox="0 0 60 60"
	class={cls}
	aria-hidden="true"
>
	<!-- body -->
	<rect x="14" y="32" width="30" height="15" rx="5" fill="currentColor"/>
	<!-- roof -->
	<rect x="8" y="5" width="45" height="25" rx="5" fill="currentColor" transform="rotate(-6, 30, 33)"/>
</svg>
