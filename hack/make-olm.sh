#!/bin/bash

PROJECT_ROOT="$(readlink -e $(dirname "$BASH_SOURCE[0]")/../)"
CSV_VERSION="${CSV_VERSION:-0.0.0}"
CATALOG_DIR="${PROJECT_ROOT}/deploy/olm-catalog"
PACKAGE_DIR="${CATALOG_DIR}/kubevirt-web-ui-operator"
BUNDLE_DIR="${PACKAGE_DIR}/${CSV_VERSION}"

which operator-sdk &> /dev/null || {
	echo "operator-sdk not found (see https://github.com/operator-framework/operator-sdk)"
	exit 1
}

which operator-courier &> /dev/null || {
	echo "operator-courier not found (see https://github.com/operator-framework/operator-courier)"
	exit 2
}

operator-sdk olm-catalog gen-csv --csv-version ${CSV_VERSION} --update-crds

# Create package directory if it doesn't exist
mkdir -p "${PACKAGE_DIR}"
rm -rf "${BUNDLE_DIR}"
cat << EOF > ${PACKAGE_DIR}/kubevirt-web-ui-package.yaml
packageName: kubevirt-web-ui
channels:
  - name: alpha
    currentCSV: kubevirt-web-ui-operator.v${CSV_VERSION}
EOF

# Correct operator-sdk naming
mv "${CATALOG_DIR}/web-ui-operator/${CSV_VERSION}" "${BUNDLE_DIR}"
rmdir "${CATALOG_DIR}/web-ui-operator"

python3 "${PROJECT_ROOT}/hack/update-olm.py" \
	"${BUNDLE_DIR}/web-ui-operator.v${CSV_VERSION}.clusterserviceversion.yaml" > \
	"${BUNDLE_DIR}/kubevirt-web-ui-operator.v${CSV_VERSION}.clusterserviceversion.yaml"
rm "${BUNDLE_DIR}/web-ui-operator.v${CSV_VERSION}.clusterserviceversion.yaml"

sed -i "s|latest|v${CSV_VERSION}|g" "${BUNDLE_DIR}/kubevirt-web-ui-operator.v${CSV_VERSION}.clusterserviceversion.yaml"

cp "${PACKAGE_DIR}/kubevirt-web-ui-package.yaml" "${BUNDLE_DIR}/kubevirt-web-ui-package.yaml"
operator-courier verify --ui_validate_io ${BUNDLE_DIR} && echo "OLM verify passed" || echo "OLM verify failed"
rm "${BUNDLE_DIR}/kubevirt-web-ui-package.yaml"
