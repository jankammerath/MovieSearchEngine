document.addEventListener('DOMContentLoaded', () => {
    const genreSelect = document.getElementById('genre');
    const startYearSelect = document.getElementById('startYear');
    const endYearSelect = document.getElementById('endYear');
    const searchBtn = document.getElementById('searchBtn');
    const movieGrid = document.getElementById('movieGrid');
    const resultsCount = document.getElementById('resultsCount');
    const prevBtn = document.getElementById('prevBtn');
    const nextBtn = document.getElementById('nextBtn');
    const pageInfo = document.getElementById('pageInfo');

    let currentOffset = 0;
    const limit = 32;

    // Load filter options on startup
    fetchOptions();

    // Event listeners
    searchBtn.addEventListener('click', () => {
        currentOffset = 0; // Reset pagination
        performSearch();
    });

    prevBtn.addEventListener('click', () => {
        if (currentOffset >= limit) {
            currentOffset -= limit;
            performSearch();
        }
    });

    nextBtn.addEventListener('click', () => {
        currentOffset += limit;
        performSearch();
    });

    async function fetchOptions() {
        try {
            // Fetch Genres
            const genreRes = await fetch('/api/genre');
            let genres = await genreRes.json();
            genres.sort();
            genres.forEach(g => {
                if (g) {
                    const opt = document.createElement('option');
                    opt.value = g;
                    opt.textContent = g;
                    genreSelect.appendChild(opt);
                }
            });

            // Fetch Years
            const yearRes = await fetch('/api/year');
            let years = await yearRes.json();
            years.sort((a, b) => b - a); // Sort descending (newest first)
            years.forEach(y => {
                const optStart = document.createElement('option');
                optStart.value = y;
                optStart.textContent = y;
                startYearSelect.appendChild(optStart);

                const optEnd = document.createElement('option');
                optEnd.value = y;
                optEnd.textContent = y;
                endYearSelect.appendChild(optEnd);
            });

            // Run initial search
            performSearch();

        } catch (error) {
            console.error("Failed to load options:", error);
            movieGrid.innerHTML = '<p style="color:red;">Error loading initial data. Is the server running?</p>';
        }
    }

    async function performSearch() {
        const genre = genreSelect.value;
        const startYear = startYearSelect.value;
        const endYear = endYearSelect.value;

        // Build query string
        const params = new URLSearchParams({
            offset: currentOffset,
            limit: limit
        });
        
        if (genre) params.append('genre', genre);
        if (startYear) params.append('startYear', startYear);
        if (endYear) params.append('endYear', endYear);

        movieGrid.innerHTML = 'Loading...';

        try {
            const res = await fetch(`/api/search?${params.toString()}`);
            if (!res.ok) throw new Error("Search failed");
            
            const movies = await res.json();
            renderMovies(movies);
            
            // Manage Pagination state
            const page = Math.floor(currentOffset / limit) + 1;
            pageInfo.textContent = `Page ${page}`;
            
            prevBtn.disabled = currentOffset === 0;
            // If we got fewer results than limit, we hit the end
            nextBtn.disabled = movies.length < limit;

        } catch (error) {
            console.error("Search error:", error);
            movieGrid.innerHTML = '<p style="color:red;">Error conducting search.</p>';
        }
    }

    function renderMovies(movies) {
        movieGrid.innerHTML = '';
        resultsCount.textContent = movies ? movies.length : 0;

        if (!movies || movies.length === 0) {
            movieGrid.innerHTML = '<p>No movies found matching your criteria.</p>';
            return;
        }

        movies.forEach(m => {
            const card = document.createElement('div');
            card.className = 'movie-card';

            const poster = document.createElement('img');
            poster.className = 'movie-poster';
            poster.src = `/media/${m.movieId}`;
            poster.alt = `${m.title} poster`;
            poster.onerror = () => { 
                poster.src = 'https://via.placeholder.com/200x300?text=No+Poster'; 
            };

            const title = document.createElement('div');
            title.className = 'movie-title';
            title.textContent = m.title;

            const info = document.createElement('div');
            info.className = 'movie-info';
            info.textContent = `Year: ${m.year === 0 ? 'Unknown' : m.year}`;

            const idInfo = document.createElement('div');
            idInfo.className = 'movie-info';
            idInfo.textContent = `ID: ${m.movieId}`;

            card.appendChild(poster);
            card.appendChild(title);
            card.appendChild(info);
            card.appendChild(idInfo);

            if (m.genres && m.genres.length > 0) {
                m.genres.forEach(g => {
                    const badge = document.createElement('span');
                    badge.className = 'genre-tag';
                    badge.textContent = g;
                    card.appendChild(badge);
                });
            }

            movieGrid.appendChild(card);
        });
    }
});