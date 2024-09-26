import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter, type RouteLocation } from 'vue-router'
export function useSearchParams(queryMap: Map<string, string | Array<string>>) {
  const router = useRouter();
	const route = useRoute();
  const searchParams = ref(queryMap);
  
  const updateFilters = () => {
		if (searchParams.value.has('page'))
    	searchParams.value.set('page', '1');
    updateQueryParams();
  };
  
  const removeFilter = (filter: {key: string, value: string}) => {
		if (!searchParams.value.has(filter.key)) return;
		if (Array.isArray(searchParams.value.get(filter.key))) {
			let newArray = searchParams.value.get(filter.key) as string[];
			searchParams.value.set(filter.key, newArray.filter((v: string) => v !== filter.value));
		} else {
			searchParams.value.set(filter.key, '');
		}
    updateFilters();
  };
  
  const resetFilters = () => {
		searchParams.value.forEach((value, key) => {
			if (Array.isArray(value)) {
				searchParams.value.set(key, []);
			} else {
				searchParams.value.set(key, '');
			}
		});
    updateFilters();
  };
  
  const updateQueryParams = () => {
		const query:Record<string, string | Array<string>> = {};
		searchParams.value.forEach((value, key) => {
			if (key === 'page') {
				const page = parseInt(value as string);
				if (page > 1) {
					query[key] = value;
				}
				return;
			}
			if (Array.isArray(value)) {
				if (value.length > 0) {
					query[key] = value.join(',');
				}
			} else {
				if (value) {
					query[key] = value;
				}
			}
		});
    router.replace({
      query,
    });
  };
  
  const initializeFromQueryParams = () => {
		searchParams.value.forEach((value, key) => {
			if (key === 'page') {
				searchParams.value.set(key, (route.query[key] || '1') as string);
				return;
			}
			if (Array.isArray(value)) {
				let qstring = route.query[key] as string;
				searchParams.value.set(key, []);
				if (qstring === undefined) return;
				searchParams.value.set(key, (qstring.split(',') || []) as string[]);
			} else {
				searchParams.value.set(key, (route.query[key] || '') as string);
			}
		});
  };

	const setSearchParams = (key: string, value: string | string[]) => {
		searchParams.value.set(key, value);
		updateFilters();
	};
  
  onMounted(() => {
    initializeFromQueryParams();
  });
  
  watch(() => route.query, () => {
	console.log("watching  ",route.query);

    initializeFromQueryParams();
  }, { deep: true });

  const hasActiveFilters = computed(() => {
	for (const [key, value] of searchParams.value) {
		if (key === 'page') continue;
	if (value.length > 0) {
	  return true;
		}
		}
	return false;  
	});

	return {
		searchParams,
		hasActiveFilters,
		setSearchParams,
		removeFilter,
		resetFilters,
	};
}