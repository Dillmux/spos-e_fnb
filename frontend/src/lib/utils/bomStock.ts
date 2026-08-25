import type { CartItem, Product, ProductRecipe, Ingredient } from '$lib/types';

export interface AggregateBOMUsage {
  ingredientId: string;
  name: string;
  unit: string;
  currentStock: number;
  minStockAlert: number;
  totalRequiredInCart: number;
  remainingStockAfterCart: number;
}

/**
 * Calculates total aggregate raw material usage across all items currently in the cart.
 */
export function calculateAggregateBOMStock(cartItems: CartItem[]): Map<string, AggregateBOMUsage> {
  const usageMap = new Map<string, AggregateBOMUsage>();

  for (const item of cartItems) {
    const recipes: ProductRecipe[] = item.product?.recipes || (item.product as any)?.Recipes || (item.product as any)?.recipe || [];
    for (const r of recipes) {
      const ing: Ingredient | undefined = r.Ingredient || (r as any).ingredient;
      const ingId = r.ingredient_id || ing?.id;
      if (!ingId) continue;

      const currentStock = ing?.current_stock ?? ing?.stock ?? 0;
      const minAlert = ing?.min_stock_alert ?? ing?.min_stock ?? 0;
      const ingName = ing?.name || r.ingredient_name || 'Bahan Baku';
      const unit = ing?.unit || r.unit || 'unit';
      const qtyReq = r.quantity_required ?? (r as any).QuantityRequired ?? r.quantity ?? 1;
      const itemTotalReq = item.quantity * qtyReq;

      if (!usageMap.has(ingId)) {
        usageMap.set(ingId, {
          ingredientId: ingId,
          name: ingName,
          unit,
          currentStock,
          minStockAlert: minAlert,
          totalRequiredInCart: itemTotalReq,
          remainingStockAfterCart: currentStock - itemTotalReq,
        });
      } else {
        const existing = usageMap.get(ingId)!;
        existing.totalRequiredInCart += itemTotalReq;
        existing.remainingStockAfterCart = existing.currentStock - existing.totalRequiredInCart;
      }
    }
  }

  return usageMap;
}

/**
 * Helper to get a formatted array of aggregate BOM ingredient requirements for display in CartPanel or PaymentModal.
 */
export function getCartBOMSummaryList(cartItems: CartItem[]): AggregateBOMUsage[] {
  const usageMap = calculateAggregateBOMStock(cartItems);
  return Array.from(usageMap.values()).sort((a, b) => a.name.localeCompare(b.name));
}

import { get } from 'svelte/store';
import { featureFlags } from '$lib/stores/config';

/**
 * Validates whether adding `qtyToAdd` of `product` to `cartItems` is allowed based on aggregate BOM stock.
 */
export function validateBOMStockAddition(
  product: Product,
  qtyToAdd: number,
  cartItems: CartItem[]
): {
  allowed: boolean;
  reason?: string;
  maxOrderablePortions?: number;
  depletedIngredientName?: string;
} {
  if (!product.is_available) {
    return {
      allowed: false,
      reason: `Menu '${product.name}' sedang dinonaktifkan / tidak dapat dipesan.`,
      maxOrderablePortions: 0,
    };
  }

  const flags = get(featureFlags);
  if (!flags.enable_bom) {
    return { allowed: true };
  }

  const recipes: ProductRecipe[] = product.recipes || (product as any)?.Recipes || (product as any)?.recipe || [];
  if (recipes.length === 0) {
    return { allowed: true };
  }

  const currentUsageMap = calculateAggregateBOMStock(cartItems);

  let minAllowedPortions = Infinity;
  let restrictingIngName = '';
  let restrictingStock = 0;
  let restrictingUsedInCart = 0;
  let restrictingUnit = '';

  for (const r of recipes) {
    const ing: Ingredient | undefined = r.Ingredient || (r as any).ingredient;
    const ingId = r.ingredient_id || ing?.id;
    const ingName = ing?.name || r.ingredient_name || 'Bahan Baku';
    const currentStock = ing?.current_stock ?? ing?.stock ?? 0;
    const qtyReq = r.quantity_required ?? (r as any).QuantityRequired ?? r.quantity ?? 1;
    const unit = ing?.unit || r.unit || 'unit';

    const usedInCart = ingId ? (currentUsageMap.get(ingId)?.totalRequiredInCart || 0) : 0;
    const remainingStock = currentStock - usedInCart;

    const possiblePortions = Math.max(0, Math.floor(remainingStock / qtyReq));

    if (possiblePortions < minAllowedPortions) {
      minAllowedPortions = possiblePortions;
      restrictingIngName = ingName;
      restrictingStock = currentStock;
      restrictingUsedInCart = usedInCart;
      restrictingUnit = unit;
    }
  }

  if (minAllowedPortions < qtyToAdd) {
    const remainingForThisItem = Math.max(0, restrictingStock - restrictingUsedInCart);
    const reason = minAllowedPortions <= 0
      ? `Stok bahan baku '${restrictingIngName}' tidak mencukupi untuk ditambah ke keranjang! (Total Stok: ${restrictingStock} ${restrictingUnit}, Terpakai di Keranjang: ${restrictingUsedInCart} ${restrictingUnit}, Sisa: ${remainingForThisItem} ${restrictingUnit}).`
      : `Batas stok bahan baku '${restrictingIngName}' hanya menyisakan maksimal ${minAllowedPortions} porsi untuk menu '${product.name}'.`;

    return {
      allowed: false,
      reason,
      maxOrderablePortions: minAllowedPortions,
      depletedIngredientName: restrictingIngName,
    };
  }

  return { allowed: true, maxOrderablePortions: minAllowedPortions };
}
