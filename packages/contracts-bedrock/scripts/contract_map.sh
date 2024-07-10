#!/bin/bash
set -o pipefail

L1_URL="${1:?Must specify L1 RPC URL}"
L1_ADDRESSES="${2:?Must specify L1 addresses json}"
OUTPUT="${3:-relations}"

addresses=$(jq -r '.[]' "$L1_ADDRESSES")

contract_addresses=()
processed_addresses=()
dots=()

while IFS= read -r address; do
    contract_addresses+=("$address")
done <<< "$addresses"

address_exists() {
    local addr="$1"
    for processed_addr in "${processed_addresses[@]}"; do
        if [[ "$processed_addr" == "$addr" ]]; then
            return 0
        fi
    done
    return 1
}

check_admin() {
    local addr="$1"
    admin=$(cast adm "$addr" --rpc-url "$L1_URL")

    if [[ $? == 0 && "$admin" != "0x0000000000000000000000000000000000000000" ]]; then
        contract_addresses+=( "$admin" )
        echo "   -> Admin: $admin"
        add_relation "$admin" "$addr" "admin"
    fi

    return 0
}

check_owners() {
    local addr="$1"

    # suppressing stderr (and unset -e) as failure is expected when this abi does not exist
    # getOwners defined in OwnerManager on GnosisSafe contract
    if owners=$(cast call "$addr" --rpc-url "$L1_URL" 'getOwners()(address[])' 2>/dev/null) ; then
        # trim pseudo json output
        tr=$(echo "$owners" | tr -d '[],')
        owners_arr=( "$tr" )

        # Iterate over the values
        for owner in "${owners_arr[@]}"; do
            echo "   -> Multisig Owner: $owner"
            add_relation "$owner" "$addr" "multisig_owner"
            contract_addresses+=( "$owner" )
        done
    fi

    # owner defined in Ownable on OpenZeppelin abstract contract
    if owner=$(cast call "$addr" --rpc-url "$L1_URL" 'owner()(address)' 2>/dev/null) ; then
        echo "   -> Owner: $owner"
        add_relation "$owner" "$addr" "owner"
        contract_addresses+=( "$owner" )
    fi

    return 0
}

check_implementation() {
    local addr="$1"

    impl=$(cast implementation "$addr" --rpc-url "$L1_URL")

    if [[ $? == 0 && "$impl" != "0x0000000000000000000000000000000000000000" ]]; then
        contract_addresses+=( "$impl" )
        echo "   -> Impl: $impl"
        add_relation "$addr" "$impl" "proxies"
    fi

    return 0
}

get_name() {
    local addr
    local result

    addr=$(cast to-check-sum-address "$1")
    result=$(jq -r "to_entries | map(select(.value == \"$addr\")) | .[0].key" "$L1_ADDRESSES")

    if [[ ${#result} -gt 4 ]]; then
        printf "%s\n(%s)" "$addr" "$result"
    else
        echo "$addr"
    fi
}

add_relation() {
    local source="$1"
    local destination="$2"
    local label="$3"

    local source_name
    local destination_name

    source_name=$(get_name "$source")
    destination_name=$(get_name "$destination")

    dots+=("\"$source_name\" -> \"$destination_name\"[label = \"$label\"];")
}

# while loop to allow for modification of the array during iteration
i=0
while [ $i -lt ${#contract_addresses[@]} ]; do
    address="$(cast to-check-sum-address "${contract_addresses[$i]}")"
    if address_exists "$address"; then
        # already processed this address, skip iteration
        i=$((i + 1))
        continue
    fi

    echo "Checking $address"

    check_admin "$address"
    check_owners "$address"
    check_implementation "$address"

    processed_addresses+=("$address")
    i=$((i + 1))
done

# write out chart
echo "digraph {" > "$OUTPUT".dot
echo "rankdir=\"LR\";" >> "$OUTPUT".dot
for dot in "${dots[@]}"; do
    echo "$dot" >> "$OUTPUT".dot
done
echo "}" >> "$OUTPUT".dot

dot "$OUTPUT".dot -Tpng -o "$OUTPUT".png
open "$OUTPUT".png
