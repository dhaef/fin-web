const form = document.getElementById('category-form');

form.addEventListener('submit', async (e) => {
  e.preventDefault();

  // clear errors on submit
  const errElements = ['label', 'priority', 'values'];
  for (const i of errElements) {
    const el = document.querySelector(`.${i}-err`);
    if (el) {
      el.textContent = '';
    }
  }

  const label = form.querySelector('input[name="label"]')?.value;
  const priority = form.querySelector('input[name="priority"]')?.value;
  const values = [];

  for (const v of form.querySelectorAll('.value')) {
    if (!v.value) continue;

    values.push({
      id: v.id ? v.id : undefined,
      value: v.value,
    });
  }

  const formData = new FormData();
  formData.append('label', label);
  formData.append('priority', priority);
  formData.append('values', JSON.stringify(values));

  const resp = await fetch(document.location.pathname, {
    method: 'POST',
    body: formData,
  });
  const data = await resp.json();

  if (resp.status === 201) {
    window.location.href = data.redirect;
    return;
  }

  if (data.errs) {
    for (const [key, err] of Object.entries(data.errs)) {
      const el = document.querySelector(`.${key}-err`);
      if (el) {
        el.textContent = err;
      }
    }
  }
});

const values = document.getElementById('values');
const addValueBtn = document.getElementById('add-value');
addValueBtn.addEventListener('click', (e) => {
  e.preventDefault();

  const parentEl = document.createElement('div');
  const el = document.createElement('input');
  el.className = 'value';
  const removeBtn = document.createElement('button');
  removeBtn.className = 'rm-value-btn';
  removeBtn.innerText = '🗑️';
  removeBtn.addEventListener('click', removeValueButton);

  parentEl.append(el);
  parentEl.append(removeBtn);
  values.append(parentEl);
});

function removeValueButton(e) {
  e.preventDefault();
  e.target.parentElement.remove();
}

const removeValueButtons = document.querySelectorAll('.rm-value-btn');
for (const btn of removeValueButtons) {
  btn.addEventListener('click', removeValueButton);
}
