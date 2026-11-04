"""Build data/seasonality.json from the Census Monthly Retail Trade workbook (stdlib only).

Run:  python3 data/seasonality.py
Input: data/census/mrtssales92-present.xlsx (https://www.census.gov/retail/mrts/www/mrtssales92-present.xlsx,
downloaded Nov 3, 2026). Demand index per category = monthly NOT ADJUSTED sales ÷ that year's monthly mean,
averaged over YEARS (2023-2025, the latest full years). Jewelry stores (44831) are suppressed by Census from
March 2021 on, so that one category uses 2017-2019 (the latest jewelry-specific pattern; its parent 448 shows
the same shape in 2023-2025). Ad cost is one table for every category (Gupta Media Meta CPM 2024-2025).
"""
import json
import os
import xml.etree.ElementTree as ET
import zipfile

HERE = os.path.dirname(os.path.abspath(__file__))
XLSX = os.path.join(HERE, "census", "mrtssales92-present.xlsx")
OUT = os.path.join(HERE, "seasonality.json")
M = "{http://schemas.openxmlformats.org/spreadsheetml/2006/main}"
YEARS = ["2023", "2024", "2025"]
CATEGORIES = [  # key, Setup label, Census code, years
    ("jewelry", "Jewelry", "44831", ["2017", "2018", "2019"]),
    ("apparel", "Clothing and accessories", "448", YEARS),
    ("beauty", "Beauty, health and personal care", "446", YEARS),
    ("home", "Furniture and home", "442", YEARS),
    ("electronics", "Electronics and appliances", "443", YEARS),
    ("hobby", "Sporting goods, hobby, music and books", "451", YEARS),
    ("food", "Food and drink", "445", YEARS),
    ("online", "Online store, other goods", "454", YEARS),
]
AD_COST = {"1": 0.87, "2": 0.87, "3": 0.97, "4": 0.97, "5": 1.01, "6": 0.99, "7": 0.92, "8": 0.95, "9": 1.02, "10": 0.97, "11": 1.18, "12": 1.27}
AD_SOURCE = {"for": "ad_cost", "source": "Gupta Media, Meta (Facebook + Instagram) average CPM by month, 2024 and 2025; each month ÷ that year's monthly mean, averaged over the two years. Latest full years with a monthly series; 2026 only runs to September (AdKit worldwide medians) and a partial year cannot be normalized to an annual mean",
             "url": "https://www.guptamedia.com/social-media-ads-cost", "year": 2025}


def sheet_rows(z, name):
    wb = ET.fromstring(z.read("xl/workbook.xml"))
    rels = ET.fromstring(z.read("xl/_rels/workbook.xml.rels"))
    rid = {r.get("Id"): r.get("Target") for r in rels}
    sheets = {s.get("name"): s.get("{http://schemas.openxmlformats.org/officeDocument/2006/relationships}id") for s in wb.find(M + "sheets")}
    ss = ["".join(t.text or "" for t in si.iter(M + "t")) for si in ET.fromstring(z.read("xl/sharedStrings.xml"))]
    root = ET.fromstring(z.read("xl/" + rid[sheets[name]].lstrip("/").replace("xl/", "")))
    for row in root.iter(M + "row"):
        vals = []
        for c in row.findall(M + "c"):
            v = c.find(M + "v")
            x = v.text if v is not None else ""
            vals.append(ss[int(x)] if c.get("t") == "s" else x)
        yield vals


def not_adjusted(z, year, code):
    """The first row for `code` in a year's sheet is the NOT ADJUSTED block; 12 monthly values in millions of USD."""
    for vals in sheet_rows(z, year):
        if vals and vals[0] == code:
            return [float(x) for x in vals[2:14]]
    raise SystemExit(f"{code} not in sheet {year}")


def main():
    z = zipfile.ZipFile(XLSX)
    cats = {}
    for key, label, code, years in CATEGORIES:
        idx = [0.0] * 12
        for y in years:
            m = not_adjusted(z, y, code)
            mean = sum(m) / 12
            for i in range(12):
                idx[i] += m[i] / mean / len(years)
        span = f"{years[0]}-{years[-1]}"
        src = (f"U.S. Census Bureau, Monthly Retail Trade (mrtssales92-present.xlsx, downloaded Nov 3, 2026): "
               f"{label.lower()} stores, NAICS {code}, not seasonally adjusted; monthly sales ÷ that year's monthly mean, averaged over {span}")
        if key == "jewelry":
            src += (". Census suppresses the jewelry line from March 2021 on ((S) in every later sheet), so this is the latest "
                    "jewelry-specific pattern; its parent category 448 shows the same holiday shape in 2023-2025 (December 1.56×, November 1.12-1.16×)")
        cats[key] = {"label": label, "demand": {str(i + 1): round(v, 2) for i, v in enumerate(idx)},
                     "source": {"for": "demand", "source": src, "url": "https://www.census.gov/retail/mrts/www/mrtssales92-present.xlsx", "year": int(years[-1])}}
    out = {"market": "US", "categories": cats, "ad_cost": AD_COST, "ad_cost_source": AD_SOURCE,
           "note": "Indexes are relative to an average month (mean 1.0). Applied as a prior for the month being planned; never used to calibrate or backtest, because one year of a brand's history cannot separate seasonality from everything else."}
    with open(OUT, "w") as f:
        json.dump(out, f, indent=1, ensure_ascii=False)
        f.write("\n")
    for k, c in cats.items():
        print(f"{k:12s} Nov {c['demand']['11']:.2f} Dec {c['demand']['12']:.2f} Jan {c['demand']['1']:.2f}")


if __name__ == "__main__":
    main()
