export const UUID_REGEX = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

// Kiosk basket: 1–20 lines, item_id = slot id (vending_machine_to_items.id) > 0, amount 1–10.
// Duplicate slots are summed in SQL. Returns null if anything is off.
export function parseItems(items: unknown): { slot_id: number; quantity: number }[] | null {
  if (!Array.isArray(items) || items.length === 0 || items.length > 20) return null;
  const out = [];
  for (const it of items) {
    if (!Number.isInteger(it?.item_id) || it.item_id <= 0) return null;
    if (!Number.isInteger(it?.amount) || it.amount < 1 || it.amount > 10) return null;
    out.push({ slot_id: it.item_id, quantity: it.amount });
  }
  return out;
}
