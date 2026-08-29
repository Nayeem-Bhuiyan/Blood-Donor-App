const donorListEl = document.getElementById('donor-list');
const paginationEl = document.getElementById('pagination');
const modalEl = document.getElementById('donorModal');
const modalBody = document.getElementById('modalBody');
const closeModalBtn = document.getElementById('closeModal');
const donorFormModal = document.getElementById('donorFormModal');
const donorForm = document.getElementById('donorForm');
const formTitle = document.getElementById('formTitle');
const openCreateModalBtn = document.getElementById('openCreateModal');
const closeFormModalBtn = document.getElementById('closeFormModal');
const cancelFormBtn = document.getElementById('cancelForm');
const searchInput = document.getElementById('searchInput');
const toastContainer = document.getElementById('toastContainer');
const pageSizeSelect = document.getElementById('pageSizeSelect');
const resultsSummary = document.getElementById('resultsSummary');
const resultsSummaryBottom = document.getElementById('resultsSummaryBottom');

let currentPage = 1;
let pageSize = 9;
let currentDonorDetails = null;
let allDonors = [];
let searchTerm = '';

function showToast(message, type = 'success') {
  if (!toastContainer) {
    return;
  }

  const toast = document.createElement('div');
  toast.className = `toast ${type}`;
  toast.textContent = message;

  toastContainer.appendChild(toast);

  window.setTimeout(() => {
    toast.remove();
  }, 2600);
}

function formatField(value) {
  return value || 'N/A';
}

function donorImageUrl(picture) {
  return picture ? `/uploads/${picture}` : '';
}

function formatStatusDate(date) {
  const day = String(date.getDate()).padStart(2, '0');
  const month = String(date.toLocaleString('en-US', { month: 'short' })).toLowerCase();
  const year = date.getFullYear();
  return `${day}-${month}-${year}`;
}

function donorReadyForDonation(lastDonation) {
  if (!lastDonation) {
    return 'N/A';
  }

  const donationDate = new Date(lastDonation);
  if (Number.isNaN(donationDate.getTime())) {
    return 'N/A';
  }

  const readyDate = new Date(donationDate);
  readyDate.setDate(readyDate.getDate() + 120);
  return formatStatusDate(readyDate);
}

function donorAvailabilityFromDate(lastDonation) {
  if (!lastDonation) {
    return 'Available';
  }

  const donationDate = new Date(lastDonation);
  if (Number.isNaN(donationDate.getTime())) {
    return 'Available';
  }

  const readyDate = new Date(donationDate);
  readyDate.setDate(readyDate.getDate() + 120);

  if (Date.now() >= readyDate.getTime()) {
    return 'Available';
  }

  return `Ready for donation ${formatStatusDate(readyDate)}`;
}

function renderDonorCard(donor) {
  const donorStatus = donorAvailabilityFromDate(donor.lastDonation);
  const imageHtml = donor.picture
    ? `<img class="donor-thumb" src="${donorImageUrl(donor.picture)}" alt="${donor.name} photo" />`
    : `<div class="donor-thumb placeholder">No Photo</div>`;

  return `
    <article class="donor-card">
      <div class="card-header">
        <h3 class="card-title">${donor.name}</h3>
        <span class="blood-badge">${donor.bloodGroup}</span>
      </div>

      ${imageHtml}

      <div class="card-meta">
        <div class="meta-row"><span>Location</span><span>${donor.location}</span></div>
        <div class="meta-row"><span>Age</span><span>${donor.age}</span></div>
        <div class="meta-row"><span>Gender</span><span>${donor.gender}</span></div>
      </div>

      <span class="status ${donorStatus.startsWith('Available') ? 'available' : 'busy'}">
        ${donorStatus}
      </span>

      <div class="card-actions">
        <button class="primary-btn" data-view-id="${donor.id}">View</button>
      </div>
    </article>
  `;
}

function filterDonors(donors, keyword) {
  if (!keyword) {
    return donors;
  }

  const normalized = keyword.trim().toLowerCase();
  if (!normalized) {
    return donors;
  }

  return donors.filter((donor) => {
    const searchableText = [
      donor.name,
      donor.bloodGroup,
      donor.gender,
      donor.lastDonation,
      String(donor.age),
      donor.location,
    ]
      .filter(Boolean)
      .join(' ')
      .toLowerCase();

    return searchableText.includes(normalized);
  });
}

function renderPagination(totalPages, activePage) {
  if (!paginationEl) {
    return;
  }

  const firstDisabled = activePage === 1 ? 'disabled' : '';
  const prevDisabled = activePage === 1 ? 'disabled' : '';
  const nextDisabled = activePage === totalPages ? 'disabled' : '';
  const lastDisabled = activePage === totalPages ? 'disabled' : '';

  let pages = [];

  if (totalPages <= 7) {
    pages = Array.from({ length: totalPages }, (_, index) => index + 1);
  } else if (activePage <= 4) {
    pages = [1, 2, 3, 4, 5, '...', totalPages];
  } else if (activePage >= totalPages - 3) {
    pages = [1, '...', totalPages - 4, totalPages - 3, totalPages - 2, totalPages - 1, totalPages];
  } else {
    pages = [1, '...', activePage - 1, activePage, activePage + 1, '...', totalPages];
  }

  let pagesHtml = `
    <button class="page-btn page-nav" data-page="1" ${firstDisabled}>First</button>
    <button class="page-btn page-nav" data-page="${activePage - 1}" ${prevDisabled}>&lt;</button>
  `;

  pages.forEach((page) => {
    if (page === '...') {
      pagesHtml += '<span class="page-ellipsis">...</span>';
      return;
    }

    const activeClass = Number(page) === activePage ? 'active' : '';
    pagesHtml += `
      <button class="page-btn ${activeClass}" data-page="${page}">${page}</button>
    `;
  });

  pagesHtml += `
    <button class="page-btn page-nav" data-page="${activePage + 1}" ${nextDisabled}>&gt;</button>
    <button class="page-btn page-nav" data-page="${totalPages}" ${lastDisabled}>Last</button>
  `;

  paginationEl.innerHTML = pagesHtml;

  paginationEl.querySelectorAll('.page-btn').forEach((button) => {
    button.addEventListener('click', () => {
      const targetPage = Number(button.dataset.page);
      if (!Number.isNaN(targetPage) && targetPage >= 1 && targetPage <= totalPages) {
        currentPage = targetPage;
        renderCurrentPage();
      }
    });
  });
}

function renderModal(donor) {
  const donorStatus = donorAvailabilityFromDate(donor.lastDonation);
  currentDonorDetails = donor;
  const photoHtml = donor.picture
    ? `<div class="modal-photo-wrap"><img class="modal-photo" src="${donorImageUrl(donor.picture)}" alt="${donor.name} photo" /></div>`
    : `<div class="modal-photo-wrap placeholder">No Photo</div>`;

  modalBody.innerHTML = `
    <div class="modal-header">
      <h2>${donor.name}</h2>
      <span class="blood-badge">${donor.bloodGroup}</span>
    </div>
    ${photoHtml}
    <div class="modal-body">
      <div class="modal-row"><span>Phone</span><span>${formatField(donor.phone)}</span></div>
      <div class="modal-row"><span>Email</span><span>${formatField(donor.email)}</span></div>
      <div class="modal-row"><span>Location</span><span>${formatField(donor.location)}</span></div>
      <div class="modal-row"><span>Age</span><span>${formatField(donor.age)}</span></div>
      <div class="modal-row"><span>Gender</span><span>${formatField(donor.gender)}</span></div>
      <div class="modal-row"><span>Address</span><span>${formatField(donor.address)}</span></div>
      <div class="modal-row"><span>Last Donation</span><span>${formatField(donor.lastDonation)}</span></div>
      <div class="modal-row"><span>Ready for Donation</span><span>${formatField(donor.readyForDonation || donorReadyForDonation(donor.lastDonation))}</span></div>
      <div class="modal-row"><span>Availability</span><span>${formatField(donorStatus)}</span></div>
      <div class="modal-row"><span>Notes</span><span>${formatField(donor.notes)}</span></div>
    </div>
    <div class="form-actions" style="justify-content: flex-start; margin-top: 18px;">
      <button type="button" class="primary-btn" data-edit-id="${donor.id}">Edit</button>
    </div>
  `;

  const editBtn = modalBody.querySelector('[data-edit-id]');
  if (editBtn) {
    editBtn.addEventListener('click', () => {
      openFormModal('edit', donor);
      modalEl.classList.add('hidden');
    });
  }
}

function openFormModal(mode, donor = null) {
  donorForm.reset();
  donorForm.dataset.mode = mode;
  donorForm.dataset.id = donor ? donor.id : '';

  const pictureField = donorForm.elements.namedItem('picture');
  if (pictureField) {
    pictureField.value = '';
  }

  if (mode === 'edit' && donor) {
    formTitle.textContent = 'Update Donor';
    Object.entries(donor).forEach(([key, value]) => {
      const field = donorForm.elements.namedItem(key);
      if (!field || field.name === 'picture') {
        return;
      }

      if (field.type === 'file') {
        field.value = '';
        return;
      }

      field.value = value ?? '';
    });
  } else {
    formTitle.textContent = 'Add Donor';
    donorForm.elements.namedItem('availability').value = 'Available';
    donorForm.elements.namedItem('bloodGroup').value = 'A+';
    donorForm.elements.namedItem('gender').value = 'Male';
  }

  donorFormModal.classList.remove('hidden');
  donorFormModal.setAttribute('aria-hidden', 'false');
}

function closeFormModal() {
  donorFormModal.classList.add('hidden');
  donorFormModal.setAttribute('aria-hidden', 'true');
  donorForm.reset();
}

function renderCurrentPage() {
  allDonors = allDonors.map((donor) => ({
    ...donor,
    availability: donorAvailabilityFromDate(donor.lastDonation),
  }));

  const filtered = filterDonors(allDonors, searchTerm);
  const total = filtered.length;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  if (currentPage > totalPages) {
    currentPage = totalPages;
  }

  const start = (currentPage - 1) * pageSize;
  const end = start + pageSize;
  const items = filtered.slice(start, end);

  donorListEl.innerHTML = items.map(renderDonorCard).join('');

  donorListEl.querySelectorAll('[data-view-id]').forEach((button) => {
    button.addEventListener('click', async () => {
      const donorId = Number(button.dataset.viewId);
      const donor = items.find((item) => item.id === donorId);
      if (donor) {
        renderModal(donor);
        modalEl.classList.remove('hidden');
        modalEl.setAttribute('aria-hidden', 'false');
      }
    });
  });

  if (pageSizeSelect) {
    pageSizeSelect.value = String(pageSize);
  }

  const startIndex = total === 0 ? 0 : start + 1;
  const endIndex = total === 0 ? 0 : Math.min(end, total);
  const summaryText = `Results: ${startIndex} - ${endIndex} of ${total}`;

  if (resultsSummary) {
    resultsSummary.textContent = summaryText;
  }

  if (resultsSummaryBottom) {
    resultsSummaryBottom.textContent = summaryText;
  }

  renderPagination(totalPages, currentPage);
}

async function fetchDonors() {
  try {
    const response = await fetch(`/api/donors?page=1&limit=9999`);
    const payload = await response.json();
    allDonors = payload.items || [];
    renderCurrentPage();
  } catch (error) {
    donorListEl.innerHTML = '<p>Unable to load donor data.</p>';
    showToast('Unable to load donor data.', 'warning');
    console.error('Error loading donor data:', error);
  }
}

async function submitDonorForm(event) {
  event.preventDefault();

  const formData = new FormData(donorForm);
  const age = Number(formData.get('age'));
  if (!Number.isNaN(age)) {
    formData.set('age', String(age));
  }

  const mode = donorForm.dataset.mode || 'create';
  const donorId = donorForm.dataset.id;

  const url = mode === 'edit' && donorId ? `/api/donors/${donorId}` : '/api/donors';
  const method = mode === 'edit' ? 'PUT' : 'POST';

  try {
    const response = await fetch(url, {
      method,
      body: formData,
    });

    if (!response.ok) {
      throw new Error('Unable to save donor');
    }

    closeFormModal();
    await fetchDonors();
    showToast(mode === 'edit' ? 'Donor updated successfully.' : 'Donor added successfully.', 'success');
  } catch (error) {
    showToast('Failed to save donor information. Please try again.', 'warning');
    console.error(error);
  }
}

closeModalBtn.addEventListener('click', () => {
  modalEl.classList.add('hidden');
  modalEl.setAttribute('aria-hidden', 'true');
});

modalEl.addEventListener('click', (event) => {
  if (event.target === modalEl) {
    modalEl.classList.add('hidden');
    modalEl.setAttribute('aria-hidden', 'true');
  }
});

openCreateModalBtn.addEventListener('click', () => {
  openFormModal('create');
});

closeFormModalBtn.addEventListener('click', closeFormModal);
cancelFormBtn.addEventListener('click', closeFormModal);
donorForm.addEventListener('submit', submitDonorForm);

searchInput.addEventListener('input', (event) => {
  searchTerm = event.target.value;
  currentPage = 1;
  renderCurrentPage();
});

if (pageSizeSelect) {
  pageSizeSelect.addEventListener('change', (event) => {
    pageSize = Number(event.target.value) || 6;
    currentPage = 1;
    renderCurrentPage();
  });
}

donorFormModal.addEventListener('click', (event) => {
  if (event.target === donorFormModal) {
    closeFormModal();
  }
});

fetchDonors();
