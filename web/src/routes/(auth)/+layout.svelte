<script lang="ts">
	let { children } = $props();
</script>

<div class="min-h-screen flex items-center justify-center bg-background px-4">
	<div class="w-full max-w-sm">
		<h1 class="text-2xl font-bold text-center mb-8">Family Board</h1>
		{@render children()}
	</div>
</div>
