document.addEventListener("DOMContentLoaded", () => {
	if (typeof Alpine === "undefined") {
		console.error("Alpine.js is not loaded yet!");
		return;
	}

	Alpine.data("qrComponent", (initialData = {}) => {
		return {
			qrData: initialData.key || "",
			username: initialData.username || "",
			remarks: initialData.remarks || "",
			isGenerated: false,
			fontFamily: "Arial, sans-serif",
			textColor: "#FF0000",
			observer: null, // Store the observer instance

			init() {
				// Generate QR immediately if the tab is visible
				if (this.$el.parentElement && !this.$el.parentElement.classList.contains("hidden")) {
					this.generateQR();
				}
				this.setupMutationObserver();
			},

			async generateQR() {
				if (!this.qrData) {
					console.log("No QR data available.");
					return;
				}
				if (this.isGenerated) {
					console.log("QR already generated, skipping regeneration.");
					return;
				}
				if (!this.$refs.qrCodeElement) {
					console.error("QR code element not found in DOM.");
					return;
				}
				this.$refs.qrCodeElement.innerHTML = "";
				const svg = await QRCode.toString(this.qrData, { type: "svg" });
				this.$refs.qrCodeElement.innerHTML = svg;
				this.isGenerated = true;
			},

			downloadQR() {
				return new Promise((resolve, reject) => {
					const svgElement = this.$refs.qrCodeElement?.querySelector("svg");
					if (!svgElement) {
						console.error("No QR code found to download.");
						return;
					}
					try {
						const canvas = document.createElement("canvas");
						const ctx = canvas.getContext("2d");
						const canvasWidth = 1024;
						const canvasHeight = 1024;
						const padding = 40;
						canvas.width = canvasWidth;
						canvas.height = canvasHeight;

						const img = new Image();
						img.crossOrigin = "anonymous";
						const svgData = new XMLSerializer().serializeToString(svgElement);
						const svgBlob = new Blob([svgData], { type: "image/svg+xml;charset=utf-8" });
						const url = URL.createObjectURL(svgBlob);

						img.onload = () => {
							ctx.fillStyle = "white";
							ctx.fillRect(0, 0, canvasWidth, canvasHeight);
							ctx.strokeStyle = "black";
							ctx.lineWidth = 2;
							ctx.strokeRect(padding, padding, canvasWidth - padding * 2, canvasHeight - padding * 2);
							ctx.fillStyle = this.textColor;
							ctx.textAlign = "center";
							const fontSizeValue = 36;
							ctx.font = `${fontSizeValue}px ${this.fontFamily}`;

							const qrSize = 750;
							const qrX = (canvasWidth - qrSize) / 2;
							const contentHeight = qrSize + fontSizeValue * 2;
							const topSpace = (canvasHeight - contentHeight) / 2;

							const username = this.username || "User";
							const textY = topSpace + fontSizeValue;
							ctx.fillText(username, canvasWidth / 2, textY);

							const qrY = textY + 20;
							ctx.drawImage(img, qrX, qrY, qrSize, qrSize);

							const remarks = this.remarks || "";
							const bottomTextY = qrY + qrSize + fontSizeValue;
							ctx.fillText(remarks, canvasWidth / 2, bottomTextY);

							canvas.toBlob((blob) => {
								try {
									const downloadUrl = URL.createObjectURL(blob);
									const downloadLink = document.createElement("a");
									downloadLink.href = downloadUrl;
									const safeUsername = (username || "user").replace(/[^a-z0-9]/gi, "_").toLowerCase();
									downloadLink.download = `qr_code_${safeUsername}.png`;
									document.body.appendChild(downloadLink);
									downloadLink.click();
									document.body.removeChild(downloadLink);
									URL.revokeObjectURL(downloadUrl);
									URL.revokeObjectURL(url);
									resolve();
								} catch (e) {
									console.error("Download error:", e);
									reject(e);
								}
							}, "image/png");
						};

						img.onerror = (e) => {
							URL.revokeObjectURL(url);
							console.error("Failed to load QR code image:", e);
							reject(e);
						};

						img.src = url;
					} catch (e) {
						console.error("Error in downloadQR:", e);
						reject(e);
					}
				});
			},

			setupMutationObserver() {
				// Clean up any existing observer
				if (this.observer) {
					this.observer.disconnect();
					console.log("Cleaned up previous MutationObserver.");
				}

				// Check if parentElement exists before observing
				if (!this.$el.parentElement) {
					console.warn("No parent element found for MutationObserver.");
					return;
				}

				this.observer = new MutationObserver((mutations) => {
					mutations.forEach((mutation) => {
						if (mutation.attributeName === "class" && this.$el.parentElement) {
							const isHidden = this.$el.parentElement.classList.contains("hidden");
							if (!isHidden && !this.isGenerated) {
								this.generateQR();
							}
						}
					});
				});

				this.observer.observe(this.$el.parentElement, { attributes: true, attributeFilter: ["class"] });
			},

			destroy() {
				// Cleanup observer when component is destroyed
				if (this.observer) {
					this.observer.disconnect();
				}
			},
		};
	});

	// Re-initialize Alpine.js after HTMX swaps
	document.body.addEventListener("htmx:afterSwap", (event) => {
		const swappedElement = event.target;
		if (swappedElement && swappedElement.querySelector("[x-data]")) {
			console.log("HTMX swap detected, re-initializing Alpine.js...");
			Alpine.initTree(swappedElement);
		}
	});
});

document.body.addEventListener('htmx:responseError', function(evt) {
	const status = evt.detail.xhr.status;
	const responseText = evt.detail.xhr.responseText;

	if (status >= 500 && status <= 599) {
		alert("Internal Server Error.\nPlease try again later or try refreshing the page\nContact support if the problem persists.");
	} else if (status >= 400 && status <= 499) {
		alert(responseText);
	}
});

// Function to initialize a single dropdown
function initializeDropdown(dropdown) {
	const trigger = dropdown.querySelector('.dropdown-trigger');
	const menu = dropdown.querySelector('.dropdown-content');
	const initialPosition = menu.getAttribute('data-initial-position') || 'bottom';
	const initialAlignment = menu.getAttribute('data-initial-alignment') || 'start';
	const dropdownId = dropdown.getAttribute('data-id');

	// Track the currently open dropdown globally
	let currentOpenDropdown = null;

	// Reset the menu position to its initial state
	function resetPosition() {
		menu.classList.remove(
			'bottom-full', 'mb-2', 'top-full', 'mt-2',
			'left-full', 'ml-2', 'right-full', 'mr-2',
			'left-0', 'right-0', 'translate-x-0', 'translate-x-center', '-translate-x-full',
			'left-1/2', '-translate-x-1/2'
		);

		switch (initialPosition) {
			case 'top':
				menu.classList.add('bottom-full', 'mb-2');
				break;
			case 'bottom':
				menu.classList.add('top-full', 'mt-2');
				break;
			case 'left':
				menu.classList.add('right-full', 'mr-2');
				menu.style.top = '0';
				break;
			case 'right':
				menu.classList.add('left-full', 'ml-2');
				menu.style.top = '0';
				break;
			default:
				menu.classList.add('top-full', 'mt-2');
		}

		if (initialPosition === 'top' || initialPosition === 'bottom') {
			switch (initialAlignment) {
				case 'start':
					menu.classList.add('left-0');
					break;
				case 'end':
					menu.classList.add('right-0');
					break;
				case 'center':
					menu.classList.add('left-1/2', '-translate-x-1/2');
					break;
				default:
					menu.classList.add('left-0');
			}
		}
	}

	function isMenuOpen() {
		return !menu.classList.contains('hidden');
	}

	function openMenu() {
		if (currentOpenDropdown && currentOpenDropdown !== dropdown) {
			const otherMenu = currentOpenDropdown.querySelector('.dropdown-content');
			otherMenu.classList.add('hidden');
		}

		resetPosition();
		menu.classList.remove('hidden');
		adjustPosition();

		currentOpenDropdown = dropdown;

		dropdown.dispatchEvent(new CustomEvent('dropdown:open', {
			detail: { dropdownId }
		}));
	}

	function closeMenu() {
		if (!isMenuOpen()) return;

		menu.classList.add('hidden');
		if (currentOpenDropdown === dropdown) {
			currentOpenDropdown = null;
		}

		dropdown.dispatchEvent(new CustomEvent('dropdown:close', {
			detail: { dropdownId }
		}));
	}

	function toggleMenu() {
		if (isMenuOpen()) {
			closeMenu();
		} else {
			openMenu();
		}
	}

	function adjustPosition() {
		const rect = menu.getBoundingClientRect();
		const viewportHeight = window.innerHeight;
		const viewportWidth = window.innerWidth;

		if (rect.bottom > viewportHeight && initialPosition === 'bottom') {
			menu.classList.remove('top-full', 'mt-2');
			menu.classList.add('bottom-full', 'mb-2');
		} else if (rect.top < 0 && initialPosition === 'top') {
			menu.classList.remove('bottom-full', 'mb-2');
			menu.classList.add('top-full', 'mt-2');
		}

		if (rect.right > viewportWidth) {
			menu.classList.remove('left-0', 'left-1/2', '-translate-x-1/2');
			menu.classList.add('right-0');
		} else if (rect.left < 0) {
			menu.classList.remove('right-0', 'left-1/2', '-translate-x-1/2');
			menu.classList.add('left-0');
		}
	}

	// Remove existing listeners to avoid duplicates
	trigger.removeEventListener('click', toggleMenu);
	trigger.addEventListener('click', (e) => {
		e.stopPropagation();
		toggleMenu();
	});

	// Close dropdown when clicking outside
	document.removeEventListener('click', outsideClickHandler);
	document.addEventListener('click', outsideClickHandler);

	function outsideClickHandler(e) {
		if (currentOpenDropdown === dropdown && !dropdown.contains(e.target)) {
			closeMenu();
		}
	}
}

// Initialize all dropdowns on page load
window.loadDropDownMenu = function() {
	const dropdowns = document.querySelectorAll('.dropdown-menu');
	dropdowns.forEach(dropdown => {
		initializeDropdown(dropdown);
	});
};

document.addEventListener('DOMContentLoaded', () => {
	window.loadDropDownMenu();
	searchInput = document.querySelector("#searchInput");
	if (searchInput) {
		searchInput.addEventListener("keyup", () => window.initUserSearch(searchInput));
	}

	sortBySelect = document.querySelector("#sortByInput")
	if (sortBySelect) {
		sortBySelect.addEventListener("change", () => window.sortUsers(sortBySelect));
	}
});

// Handle HTMX swap
function handleAfterSettle(event) {
	const swappedElement = event.detail.elt;

	if (
		(swappedElement?.getAttribute('newly-swapped') === 'true' && swappedElement?.classList?.contains('user-row')) ||
		swappedElement?.classList?.contains('users')
	) {
		// Only initialize the newly swapped dropdowns
		const newDropdowns = swappedElement.querySelectorAll('.dropdown-menu');
		newDropdowns.forEach(dropdown => {
			initializeDropdown(dropdown);
		});
	}
}

document.addEventListener('htmx:afterSettle', handleAfterSettle);

// Sorting function
window.sortUsers = function(selectElement) {
	const sortBy = selectElement.value;
	if (!sortBy) return;

	// Sort desktop table rows
	const desktopTable = document.querySelector("#desktopTable tbody");
	if (desktopTable) {
		const tableRows = Array.from(desktopTable.querySelectorAll("tr.user-row"));

		tableRows.sort((a, b) => {
			let valueA, valueB;

			switch (sortBy) {
				case '1':
					valueA = a.getAttribute("data-username").toLowerCase();
					valueB = b.getAttribute("data-username").toLowerCase();
					const lastPartA = valueA.split('/').pop();
					const lastPartB = valueB.split('/').pop();
					return lastPartA.localeCompare(lastPartB);

				case '2':
					valueA = new Date(a.getAttribute("data-start"));
					valueB = new Date(b.getAttribute("data-start"));
					return valueA - valueB;

				case '3':
					valueA = new Date(a.getAttribute("data-end"));
					valueB = new Date(b.getAttribute("data-end"));
					return valueA - valueB;

				default:
					return 0;
			}
		});

		// Reappend sorted rows to the table
		tableRows.forEach(row => desktopTable.appendChild(row));
	}

	// Sort mobile cards
	const mobileView = document.querySelector("#mobileView");
	if (mobileView) {
		const mobileCards = Array.from(mobileView.querySelectorAll(".user-card"));

		mobileCards.sort((a, b) => {
			let valueA, valueB;

			switch (sortBy) {
				case '1':
					valueA = a.getAttribute("data-username").toLowerCase();
					valueB = b.getAttribute("data-username").toLowerCase();
					const lastPartA = valueA.split('/').pop();
					const lastPartB = valueB.split('/').pop();
					return lastPartA.localeCompare(lastPartB);

				case '2':
					valueA = new Date(a.getAttribute("data-start"));
					valueB = new Date(b.getAttribute("data-start"));
					return valueA - valueB;

				case '3':
					valueA = new Date(a.getAttribute("data-end"));
					valueB = new Date(b.getAttribute("data-end"));
					return valueA - valueB;

				default:
					return 0;
			}
		});

		// Reappend sorted cards to the mobile view
		mobileCards.forEach(card => mobileView.appendChild(card));
	}
};

window.initUserSearch = function(inputElement) {
	const filter = inputElement.value.toLowerCase();

	// Filter desktop table rows
	const tableRows = document.querySelectorAll("#desktopTable tbody tr.user-row");
	tableRows.forEach(row => {
		const username = row.getAttribute("data-username").toLowerCase();
		const device = row.getAttribute("data-device").toLowerCase();
		const server = row.getAttribute("data-server").toLowerCase();
		const startDate = row.getAttribute("data-start").toLowerCase();
		const password = row.getAttribute("data-password").toLowerCase();
		const endDate = row.getAttribute("data-end").toLowerCase();
		const desc = row.getAttribute("data-desc").toLowerCase();

		const match =
			username.includes(filter) ||
			device.includes(filter) ||
			server.includes(filter) ||
			startDate.includes(filter) ||
			password.includes(filter) ||
			desc.includes(filter) ||
			endDate.includes(filter);

		row.style.display = match ? "" : "none";
	});

	// Filter mobile cards
	const mobileCards = document.querySelectorAll("#mobileView .user-card");
	mobileCards.forEach(card => {
		const username = card.getAttribute("data-username").toLowerCase();
		const device = card.getAttribute("data-device").toLowerCase();
		const server = card.getAttribute("data-server").toLowerCase();
		const password = card.getAttribute("data-password").toLowerCase();
		const startDate = card.getAttribute("data-start").toLowerCase();
		const endDate = card.getAttribute("data-end").toLowerCase();
		const desc = card.getAttribute("data-desc").toLowerCase();

		const match =
			username.includes(filter) ||
			password.includes(filter) ||
			device.includes(filter) ||
			server.includes(filter) ||
			startDate.includes(filter) ||
			desc.includes(filter) ||
			endDate.includes(filter);

		card.style.display = match ? "" : "none";
	});
};



// reload user search input if a swap is occured in the user list.
document.addEventListener('htmx:afterSettle', (event) => {

	// reload user sort if a swap is occurred and the sorting select input is presents.
	const sortBySelect = document.querySelector("#sortByInput")
	if (sortBySelect) {
		sortBySelect.addEventListener("change", () => window.sortUsers(sortBySelect));
	}

	// Get the swapped element from the event detail
	const swappedElement = event.detail.elt;
	const inputElement = swappedElement?.querySelector("#searchInput")

	// we just have to check the user-row since the user-card will also be present if the user-row is present.
	if (swappedElement?.id === "main-content" && inputElement) {
		inputElement.addEventListener("keyup", () => window.initUserSearch(inputElement));
		document.removeEventListener('htmx:afterSettle', this);
	}
});
