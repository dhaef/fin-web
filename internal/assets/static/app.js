const filterBtn = document.getElementById('filter-transactions');

if (filterBtn) {
  filterBtn.addEventListener('click', () => {
    const startDate = document.getElementById('startDate').value;
    const endDate = document.getElementById('endDate').value;
    const sortBy = document.getElementById('sortBy').value;
    const sortDirection = document.getElementById('sortDirection').value;
    const categoryOptions = document.getElementById('categories').options;

    const categories = [];
    for (let i = 0; i < categoryOptions.length; i++) {
      if (categoryOptions[i].selected) {
        categories.push(categoryOptions[i].value);
      }
    }

    const p = new URLSearchParams(location.search);

    if (startDate) p.set('startDate', startDate);
    if (endDate) p.set('endDate', endDate);
    if (sortBy) p.set('sortBy', sortBy);
    if (sortDirection) p.set('sortDirection', sortDirection);
    if (categories.length) {
      p.set('categories', categories.join(','));
    } else {
      p.delete('categories');
    }

    window.location = `${location.origin}?${p.toString()}`;
  });
}

const formatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
});

const currencyElements = document.querySelectorAll('.currency');
for (let i = 0; i < currencyElements.length; i++) {
  currencyElements[i].textContent = formatter.format(
    currencyElements[i].textContent
  );
}
